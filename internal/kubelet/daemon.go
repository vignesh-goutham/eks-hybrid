package kubelet

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/eks-hybrid/internal/api"
	"github.com/aws/eks-hybrid/internal/daemon"
)

const KubeletDaemonName = "kubelet"

var _ daemon.Daemon = &kubelet{}

type kubelet struct {
	daemonManager daemon.DaemonManager
	// environment variables to write for kubelet
	environment map[string]string
	// kubelet config flags without leading dashes
	flags map[string]string
	// awsConfig used to make aws calls
	awsConfig aws.Config
}

func NewKubeletDaemon(daemonManager daemon.DaemonManager, awsConfig aws.Config) daemon.Daemon {
	return &kubelet{
		daemonManager: daemonManager,
		environment:   make(map[string]string),
		flags:         make(map[string]string),
		awsConfig:     awsConfig,
	}
}

func (k *kubelet) Configure(cfg *api.NodeConfig) error {
	if cfg.IsHybridNode() {
		if err := k.ensureClusterDetails(cfg); err != nil {
			return err
		}
	}
	if err := k.writeKubeletConfig(cfg); err != nil {
		return err
	}
	if err := k.writeKubeconfig(cfg); err != nil {
		return err
	}
	if err := k.writeImageCredentialProviderConfig(cfg); err != nil {
		return err
	}
	if err := writeClusterCaCert(cfg.Spec.Cluster.CertificateAuthority); err != nil {
		return err
	}
	if err := k.writeKubeletEnvironment(cfg); err != nil {
		return err
	}
	return nil
}

func (k *kubelet) EnsureRunning() error {
	err := k.daemonManager.EnableDaemon(KubeletDaemonName)
	if err != nil {
		return err
	}
	return k.daemonManager.StartDaemon(KubeletDaemonName)
}

func (k *kubelet) PostLaunch(_ *api.NodeConfig) error {
	return nil
}

func (k *kubelet) Stop() error {
	return k.daemonManager.StopDaemon(KubeletDaemonName)
}

func (k *kubelet) Name() string {
	return KubeletDaemonName
}
