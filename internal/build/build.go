// Package build implements the local Docker image build pipeline for the
// inner dev loop: compile, build, push to the preload registry, and restart workloads.
package build

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/a-cordier/sew/internal/cache"
	"github.com/a-cordier/sew/internal/config"
	"github.com/a-cordier/sew/internal/logger"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/cluster"
)

// Options configures the build pipeline.
type Options struct {
	ClusterName string
	SewHome     string
	SkipPre     bool
	NoRestart   bool
	LogWriter   io.Writer
}

// Run executes the full build pipeline for a single Build entry:
// env expansion, pre-build commands, Docker build, push to preload registry, and rollout restart.
func Run(ctx context.Context, b config.Build, opts Options) error {
	dir := expandEnv(b.Dir)
	if dir == "" {
		dir = "."
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolving build directory: %w", err)
	}

	logw := opts.LogWriter
	if logw == nil {
		logw = io.Discard
	}

	if !opts.SkipPre {
		if err := logger.WithSpinner(
			"Running pre-build commands",
			func() error {
				for _, c := range b.Pre {
					if err := runCommand(dir, c, logw); err != nil {
						return err
					}
				}
				return nil
			},
		); err != nil {
			return err
		}
	}

	if err := logger.WithSpinner(
		fmt.Sprintf("Building image %s", b.Image),
		func() error {
			return dockerBuild(ctx, dir, expandEnv(b.Context), expandEnv(b.Dockerfile), b.Image, logw)
		},
	); err != nil {
		return err
	}

	if err := logger.WithSpinner(
		"Pushing image to preload registry",
		func() error {
			if err := cache.EnsurePreloadRegistry(ctx, opts.SewHome); err != nil {
				return fmt.Errorf("ensuring preload registry: %w", err)
			}
			if err := cache.ConnectPreloadToKindNetwork(ctx); err != nil {
				return fmt.Errorf("connecting preload registry to kind network: %w", err)
			}
			return cache.PushImages(ctx, []string{b.Image})
		},
	); err != nil {
		return err
	}

	if !opts.NoRestart {
		return restartWorkloads(ctx, b.Image, opts.ClusterName)
	}
	return nil
}

func expandEnv(s string) string {
	return os.ExpandEnv(s)
}

func runCommand(dir, command string, w io.Writer) error {
	fmt.Fprintf(w, "\n  $ %s\n\n", command)
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = w
	cmd.Stderr = w
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command %q failed: %w", command, err)
	}
	fmt.Fprintln(w)
	return nil
}

func dockerBuild(ctx context.Context, dir, buildContext, dockerfile, imageName string, logw io.Writer) error {
	contextDir := dir
	if buildContext != "" {
		contextDir = filepath.Join(dir, buildContext)
	}

	dockerfilePath := filepath.Join(contextDir, "Dockerfile")
	if dockerfile != "" {
		dockerfilePath = filepath.Join(dir, dockerfile)
	}

	tarBuf, dockerfileInTar, err := createBuildContext(contextDir, dockerfilePath)
	if err != nil {
		return fmt.Errorf("creating build context: %w", err)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("creating docker client: %w", err)
	}
	defer cli.Close()

	resp, err := cli.ImageBuild(ctx, tarBuf, dockertypes.ImageBuildOptions{
		Tags:       []string{imageName},
		Dockerfile: dockerfileInTar,
		Remove:     true,
	})
	if err != nil {
		return fmt.Errorf("starting docker build: %w", err)
	}
	defer resp.Body.Close()

	return drainBuildOutput(resp.Body, logw)
}

// createBuildContext creates a tar archive of contextDir. If the Dockerfile
// lives outside contextDir it is added to the tar at a synthetic path.
func createBuildContext(contextDir, dockerfilePath string) (io.Reader, string, error) {
	rel, err := filepath.Rel(contextDir, dockerfilePath)
	insideContext := err == nil && !strings.HasPrefix(rel, "..")

	var dockerfileInTar string
	if insideContext {
		dockerfileInTar = filepath.ToSlash(rel)
	} else {
		dockerfileInTar = ".sew.Dockerfile"
	}

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	if err := addDirectoryToTar(tw, contextDir); err != nil {
		tw.Close()
		return nil, "", err
	}

	if !insideContext {
		data, err := os.ReadFile(dockerfilePath)
		if err != nil {
			tw.Close()
			return nil, "", fmt.Errorf("reading Dockerfile %s: %w", dockerfilePath, err)
		}
		if err := tw.WriteHeader(&tar.Header{
			Name: dockerfileInTar,
			Mode: 0o644,
			Size: int64(len(data)),
		}); err != nil {
			tw.Close()
			return nil, "", err
		}
		if _, err := tw.Write(data); err != nil {
			tw.Close()
			return nil, "", err
		}
	}

	if err := tw.Close(); err != nil {
		return nil, "", err
	}

	return buf, dockerfileInTar, nil
}

func addDirectoryToTar(tw *tar.Writer, dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		var link string
		if d.Type()&os.ModeSymlink != 0 {
			link, err = os.Readlink(path)
			if err != nil {
				return err
			}
		}

		header, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(rel)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}

		return nil
	})
}

func drainBuildOutput(r io.Reader, logw io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		var msg struct {
			Stream string `json:"stream"`
			Error  string `json:"error"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}
		if msg.Error != "" {
			fmt.Fprintf(logw, "ERROR: %s\n", msg.Error)
			return fmt.Errorf("docker build: %s", msg.Error)
		}
		if msg.Stream != "" {
			fmt.Fprint(logw, msg.Stream)
		}
	}
	return scanner.Err()
}

// restartWorkloads finds Deployments and StatefulSets that reference the
// given image and performs a rollout restart on each.
func restartWorkloads(ctx context.Context, imageName, clusterName string) error {
	restCfg, err := clusterRESTConfig(clusterName)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("creating kubernetes client: %w", err)
	}

	patch := []byte(fmt.Sprintf(
		`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`,
		time.Now().Format(time.RFC3339),
	))

	var restarted int

	deploys, err := clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("listing deployments: %w", err)
	}
	for i := range deploys.Items {
		d := &deploys.Items[i]
		if !podSpecReferencesImage(d.Spec.Template.Spec, imageName) {
			continue
		}
		if _, err := clientset.AppsV1().Deployments(d.Namespace).Patch(
			ctx, d.Name, k8stypes.StrategicMergePatchType, patch, metav1.PatchOptions{},
		); err != nil {
			return fmt.Errorf("restarting deployment %s/%s: %w", d.Namespace, d.Name, err)
		}
		logger.Success("Restarted deployment %s/%s", d.Namespace, d.Name)
		restarted++
	}

	stsets, err := clientset.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("listing statefulsets: %w", err)
	}
	for i := range stsets.Items {
		s := &stsets.Items[i]
		if !podSpecReferencesImage(s.Spec.Template.Spec, imageName) {
			continue
		}
		if _, err := clientset.AppsV1().StatefulSets(s.Namespace).Patch(
			ctx, s.Name, k8stypes.StrategicMergePatchType, patch, metav1.PatchOptions{},
		); err != nil {
			return fmt.Errorf("restarting statefulset %s/%s: %w", s.Namespace, s.Name, err)
		}
		logger.Success("Restarted statefulset %s/%s", s.Namespace, s.Name)
		restarted++
	}

	if restarted == 0 {
		logger.Warn("No workloads found using image %s", imageName)
	}

	return nil
}

func podSpecReferencesImage(spec v1.PodSpec, imageName string) bool {
	for _, c := range spec.Containers {
		if c.Image == imageName {
			return true
		}
	}
	for _, c := range spec.InitContainers {
		if c.Image == imageName {
			return true
		}
	}
	return false
}

func clusterRESTConfig(clusterName string) (*rest.Config, error) {
	provider := cluster.NewProvider()
	kubeConfig, err := provider.KubeConfig(clusterName, false)
	if err != nil {
		return nil, fmt.Errorf("getting kubeconfig for %q: %w", clusterName, err)
	}
	restCfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfig))
	if err != nil {
		return nil, fmt.Errorf("parsing kubeconfig: %w", err)
	}
	return restCfg, nil
}
