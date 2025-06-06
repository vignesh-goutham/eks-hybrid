package ecr

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"

	"github.com/aws/eks-hybrid/internal/aws/imds"
	"github.com/aws/eks-hybrid/internal/system"
)

const hybridServicesDomain = "amazonaws.com"

// Returns the base64 encoded authorization token string for ECR of the format "AWS:XXXXX"
func GetAuthorizationToken(awsConfig *aws.Config) (string, error) {
	ecrClient := ecr.NewFromConfig(*awsConfig)
	token, err := ecrClient.GetAuthorizationToken(context.Background(), &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", err
	}
	authData := token.AuthorizationData[0].AuthorizationToken
	return *authData, nil
}

func (r *ECRRegistry) GetSandboxImage() string {
	return r.GetImageReference("eks/pause", "3.5")
}

func GetEKSRegistry(region string) (ECRRegistry, error) {
	servicesDomain, err := imds.GetProperty(imds.ServicesDomain)
	if err != nil {
		return "", err
	}

	return getEksRegistryWithServiceDomain(region, servicesDomain)
}

func GetEKSHybridRegistry(region string) (ECRRegistry, error) {
	return getEksRegistryWithServiceDomain(region, hybridServicesDomain)
}

func getEksRegistryWithServiceDomain(region, servicesDomain string) (ECRRegistry, error) {
	account, region := getEKSRegistryCoordinates(region)
	fipsInstalled, fipsEnabled, err := system.GetFipsInfo()
	if err != nil {
		return "", err
	}
	if fipsInstalled && fipsEnabled {
		fipsRegistry := getRegistry(account, "ecr-fips", region, servicesDomain)
		if addresses, err := net.LookupHost(fipsRegistry); err != nil {
			return "", err
		} else if len(addresses) > 0 {
			return ECRRegistry(fipsRegistry), nil
		}
	}
	return ECRRegistry(getRegistry(account, "ecr", region, servicesDomain)), nil
}

type ECRRegistry string

func (r *ECRRegistry) String() string {
	return string(*r)
}

func (r *ECRRegistry) GetImageReference(repository, tag string) string {
	return fmt.Sprintf("%s/%s:%s", r.String(), repository, tag)
}

func getRegistry(accountID, ecrSubdomain, region, servicesDomain string) string {
	return fmt.Sprintf("%s.dkr.%s.%s.%s", accountID, ecrSubdomain, region, servicesDomain)
}

const nonOptInRegionAccount = "602401143452"

var accountsByRegion = map[string]string{
	"ap-northeast-1": nonOptInRegionAccount,
	"ap-northeast-2": nonOptInRegionAccount,
	"ap-northeast-3": nonOptInRegionAccount,
	"ap-south-1":     nonOptInRegionAccount,
	"ap-southeast-1": nonOptInRegionAccount,
	"ap-southeast-2": nonOptInRegionAccount,
	"ca-central-1":   nonOptInRegionAccount,
	"eu-central-1":   nonOptInRegionAccount,
	"eu-north-1":     nonOptInRegionAccount,
	"eu-west-1":      nonOptInRegionAccount,
	"eu-west-2":      nonOptInRegionAccount,
	"eu-west-3":      nonOptInRegionAccount,
	"sa-east-1":      nonOptInRegionAccount,
	"us-east-1":      nonOptInRegionAccount,
	"us-east-2":      nonOptInRegionAccount,
	"us-west-1":      nonOptInRegionAccount,
	"us-west-2":      nonOptInRegionAccount,

	"af-south-1":      "877085696533",
	"ap-east-1":       "800184023465",
	"ap-east-2":       "533267051163",
	"ap-south-2":      "900889452093",
	"ap-southeast-3":  "296578399912",
	"ap-southeast-4":  "491585149902",
	"ap-southeast-5":  "151610086707",
	"ap-southeast-7":  "121268973566",
	"ca-west-1":       "761377655185",
	"cn-north-1":      "918309763551",
	"cn-northwest-1":  "961992271922",
	"eu-central-2":    "900612956339",
	"eu-isoe-west-1":  "249663109785",
	"eu-south-1":      "590381155156",
	"eu-south-2":      "455263428931",
	"il-central-1":    "066635153087",
	"me-central-1":    "759879836304",
	"me-south-1":      "558608220178",
	"mx-central-1":    "730335286997",
	"us-gov-east-1":   "151742754352",
	"us-gov-west-1":   "013241004608",
	"us-iso-east-1":   "725322719131",
	"us-iso-west-1":   "608367168043",
	"us-isob-east-1":  "187977181151",
	"us-isof-south-1": "676585237158",
}

// getEKSRegistryCoordinates returns an AWS region and account ID for the default EKS ECR container image registry
func getEKSRegistryCoordinates(region string) (string, string) {
	inRegionRegistry, ok := accountsByRegion[region]
	if ok {
		return inRegionRegistry, region
	}
	if strings.HasPrefix(region, "us-gov-") {
		return "013241004608", "us-gov-west-1"
	} else if strings.HasPrefix(region, "cn-") {
		return "961992271922", "cn-northwest-1"
	} else if strings.HasPrefix(region, "us-iso-") {
		return "725322719131", "us-iso-east-1"
	} else if strings.HasPrefix(region, "us-isob-") {
		return "187977181151", "us-isob-east-1"
	} else if strings.HasPrefix(region, "us-isof-") {
		return "676585237158", "us-isof-south-1"
	}
	return "602401143452", "us-west-2"
}
