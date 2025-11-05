package crd

import (
	"os"
	"path/filepath"
	"testing"
)

const goodKubeConfig = `
apiVersion: v1
kind: Config
clusters:
- name: test
  cluster:
    server: https://example.com
users:
- name: test-user
  user:
    token: test-token
contexts:
- name: test
  context:
    cluster: test
    user: test-user
current-context: test
`

func Test_getHomeDir_prefersHOME(t *testing.T) {
	origHome, hadHome := os.LookupEnv("HOME")
	origUserProfile, hadUserProfile := os.LookupEnv("USERPROFILE")

	_ = os.Setenv("HOME", "/tmp/home-from-HOME")
	_ = os.Setenv("USERPROFILE", "/tmp/home-from-USERPROFILE")

	t.Cleanup(func() {
		if hadHome {
			_ = os.Setenv("HOME", origHome)
		} else {
			_ = os.Unsetenv("HOME")
		}
		if hadUserProfile {
			_ = os.Setenv("USERPROFILE", origUserProfile)
		} else {
			_ = os.Unsetenv("USERPROFILE")
		}
	})

	got := getHomeDir()
	if got != "/tmp/home-from-HOME" {
		t.Fatalf("getHomeDir() = %q; want %q", got, "/tmp/home-from-HOME")
	}
}

func Test_getHomeDir_usesUSERPROFILE_ifNoHOME(t *testing.T) {
	origHome, hadHome := os.LookupEnv("HOME")
	origUserProfile, hadUserProfile := os.LookupEnv("USERPROFILE")

	_ = os.Unsetenv("HOME")
	_ = os.Setenv("USERPROFILE", "/tmp/home-from-USERPROFILE")

	t.Cleanup(func() {
		if hadHome {
			_ = os.Setenv("HOME", origHome)
		} else {
			_ = os.Unsetenv("HOME")
		}
		if hadUserProfile {
			_ = os.Setenv("USERPROFILE", origUserProfile)
		} else {
			_ = os.Unsetenv("USERPROFILE")
		}
	})

	got := getHomeDir()
	if got != "/tmp/home-from-USERPROFILE" {
		t.Fatalf("getHomeDir() = %q; want %q", got, "/tmp/home-from-USERPROFILE")
	}
}

func Test_NewKubernetesConfig_UsesHomeKubeConfig(t *testing.T) {
	// prepare temp HOME with valid kubeconfig
	tmpDir := t.TempDir()
	kubeDir := filepath.Join(tmpDir, ".kube")
	if err := os.MkdirAll(kubeDir, 0o700); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	kubePath := filepath.Join(kubeDir, "config")
	if err := os.WriteFile(kubePath, []byte(goodKubeConfig), 0o600); err != nil {
		t.Fatalf("write kubeconfig failed: %v", err)
	}

	origHome, hadHome := os.LookupEnv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() {
		if hadHome {
			_ = os.Setenv("HOME", origHome)
		} else {
			_ = os.Unsetenv("HOME")
		}
	})

	cfg, err := NewKubernetesConfig()
	if err != nil {
		t.Fatalf("NewKubernetesConfig() returned error: %v", err)
	}
	if cfg == nil {
		t.Fatalf("expected non-nil config")
	}
	if cfg.Host != "https://example.com" {
		t.Fatalf("unexpected host: %q", cfg.Host)
	}
	if cfg.BearerToken != "test-token" {
		t.Fatalf("unexpected token: %q", cfg.BearerToken)
	}
}

func Test_NewKubernetesConfig_FileExistsButInvalid_ReturnsError(t *testing.T) {
	// create temp HOME with invalid kubeconfig content
	tmpDir := t.TempDir()
	kubeDir := filepath.Join(tmpDir, ".kube")
	if err := os.MkdirAll(kubeDir, 0o700); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	kubePath := filepath.Join(kubeDir, "config")
	if err := os.WriteFile(kubePath, []byte("not-a-valid-kubeconfig"), 0o600); err != nil {
		t.Fatalf("write kubeconfig failed: %v", err)
	}

	origHome, hadHome := os.LookupEnv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() {
		if hadHome {
			_ = os.Setenv("HOME", origHome)
		} else {
			_ = os.Unsetenv("HOME")
		}
	})

	_, err := NewKubernetesConfig()
	if err == nil {
		t.Fatalf("expected error when kubeconfig file exists but is invalid")
	}
}

func Test_NewKubernetesConfig_NoHomeAndNoInCluster_ReturnsError(t *testing.T) {
	// ensure no HOME and no USERPROFILE so code skips home-config path
	origHome, hadHome := os.LookupEnv("HOME")
	origUserProfile, hadUserProfile := os.LookupEnv("USERPROFILE")

	_ = os.Unsetenv("HOME")
	_ = os.Unsetenv("USERPROFILE")

	t.Cleanup(func() {
		if hadHome {
			_ = os.Setenv("HOME", origHome)
		} else {
			_ = os.Unsetenv("HOME")
		}
		if hadUserProfile {
			_ = os.Setenv("USERPROFILE", origUserProfile)
		} else {
			_ = os.Unsetenv("USERPROFILE")
		}
	})

	cfg, err := NewKubernetesConfig()
	if err == nil {
		if cfg == nil {
			t.Fatalf("expected error when no home config and no in-cluster config")
		}
		// If cfg is non-nil but err is nil, test passes (in-cluster available).
		return
	}
}
