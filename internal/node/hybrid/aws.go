package hybrid

import (
	"context"
	"fmt"

	"github.com/aws/eks-hybrid/internal/api"
	"github.com/aws/eks-hybrid/internal/daemon"
	"github.com/aws/eks-hybrid/internal/iamrolesanywhere"
	"github.com/aws/eks-hybrid/internal/ssm"
	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

const iamRoleAnywhereProfileName = "hybrid"

func (hnp *HybridNodeProvider) ConfigureAws(ctx context.Context) error {
	if hnp.nodeConfig.IsSSM() {
		configurator := SSMAWSConfigurator{
			Manager: hnp.daemonManager,
			Logger:  hnp.logger,
		}
		if err := configurator.Configure(ctx, hnp.nodeConfig); err != nil {
			return fmt.Errorf("configuring aws credentials with SSM: %w", err)
		}

		awsConfig, err := generateAWSConfigForSSM(ctx, hnp.nodeConfig)
		if err != nil {
			return fmt.Errorf("generating aws config for SSM: %w", err)
		}

		hnp.awsConfig = &awsConfig
	} else {
		configurator := RolesAnywhereAWSConfigurator{}
		if err := configurator.Configure(ctx, hnp.nodeConfig); err != nil {
			return fmt.Errorf("configuring aws credentials with IAM Roles Anywhere: %w", err)
		}

		awsConfig, err := LoadAWSConfigForRolesAnywhere(ctx, hnp.nodeConfig)
		if err != nil {
			return fmt.Errorf("generating aws config for IAM Roles Anywhere: %w", err)
		}

		hnp.awsConfig = &awsConfig
	}
	return nil
}

func (hnp *HybridNodeProvider) GetConfig() *aws.Config {
	return hnp.awsConfig
}

type SSMAWSConfigurator struct {
	Manager daemon.DaemonManager
	Logger  *zap.Logger
}

func (c SSMAWSConfigurator) Configure(_ context.Context, nodeConfig *api.NodeConfig) error {
	ssmDaemon := ssm.NewSsmDaemon(c.Manager, nodeConfig, c.Logger)
	if err := ssmDaemon.Configure(); err != nil {
		return err
	}
	if err := ssmDaemon.EnsureRunning(); err != nil {
		return err
	}
	if err := ssmDaemon.PostLaunch(); err != nil {
		return err
	}

	return nil
}

func generateAWSConfigForSSM(ctx context.Context, nodeConfig *api.NodeConfig) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx, config.WithRegion(nodeConfig.Spec.Cluster.Region))
}

type RolesAnywhereAWSConfigurator struct{}

func (c RolesAnywhereAWSConfigurator) Configure(_ context.Context, nodeConfig *api.NodeConfig) error {
	if err := iamrolesanywhere.WriteAWSConfig(iamrolesanywhere.AWSConfig{
		TrustAnchorARN:       nodeConfig.Spec.Hybrid.IAMRolesAnywhere.TrustAnchorARN,
		ProfileARN:           nodeConfig.Spec.Hybrid.IAMRolesAnywhere.ProfileARN,
		RoleARN:              nodeConfig.Spec.Hybrid.IAMRolesAnywhere.RoleARN,
		Region:               nodeConfig.Spec.Cluster.Region,
		NodeName:             nodeConfig.Status.Hybrid.NodeName,
		ConfigPath:           nodeConfig.Spec.Hybrid.IAMRolesAnywhere.AwsConfigPath,
		SigningHelperBinPath: iamrolesanywhere.SigningHelperBinPath,
	}); err != nil {
		return err
	}

	return nil
}

func LoadAWSConfigForRolesAnywhere(ctx context.Context, nodeConfig *api.NodeConfig) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx,
		config.WithRegion(nodeConfig.Spec.Cluster.Region),
		config.WithSharedConfigFiles([]string{nodeConfig.Spec.Hybrid.IAMRolesAnywhere.AwsConfigPath}),
		config.WithSharedConfigProfile(iamRoleAnywhereProfileName),
	)
}
