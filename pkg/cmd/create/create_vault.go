package create

import (
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/jenkins-x/jx/pkg/cmd/opts/upgrade"

	"github.com/jenkins-x/jx/pkg/cmd/helper"

	"github.com/jenkins-x/jx/pkg/cloud"
	"github.com/jenkins-x/jx/pkg/kube/serviceaccount"
	"k8s.io/client-go/kubernetes"

	"github.com/banzaicloud/bank-vaults/operator/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx/pkg/kube/cluster"
	"github.com/jenkins-x/jx/pkg/kube/services"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/jenkins-x/jx/pkg/cloud/amazon"
	awsvault "github.com/jenkins-x/jx/pkg/cloud/amazon/vault"
	gkevault "github.com/jenkins-x/jx/pkg/cloud/gke/vault"
	"github.com/jenkins-x/jx/pkg/cmd/opts"
	"github.com/jenkins-x/jx/pkg/cmd/templates"
	"github.com/jenkins-x/jx/pkg/kube"
	kubevault "github.com/jenkins-x/jx/pkg/kube/vault"
	"github.com/jenkins-x/jx/pkg/log"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/jenkins-x/jx/pkg/vault"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	exposedVaultPort = "8200"
)

const (
	autoCreateTableName = "vault-data"
)

var (
	createVaultLong = templates.LongDesc(`
		Creates a Vault using the vault-operator

        The necessary flags depends on the provider of the kubernetes cluster. 
`)

	createVaultExample = templates.Examples(`
		# Create a new vault  with name my-vault
		jx create vault my-vault

		# Create a new vault with name my-vault in namespace my-vault-namespace
		jx create vault my-vault -n my-vault-namespace
	`)
)

// CreateVaultOptions the options for the create vault command
type CreateVaultOptions struct {
	CreateOptions

	GKECreateVaultOptions
	AWSCreateVaultOptions
	ClusterName         string
	Namespace           string
	SecretsPathPrefix   string
	RecreateVaultBucket bool
	NoExposeVault       bool
	BucketName          string
	KeyringName         string
	KeyName             string
	ServiceAccountName  string

	IngressConfig kube.IngressConfig
}

// GKECreateVaultOptions the options for vault on GKE
type GKECreateVaultOptions struct {
	GKEProjectID string
	GKEZone      string
}

// AWSCreateVaultOptions are the AWS specific Vault creation options
type AWSCreateVaultOptions struct {
	kubevault.AWSConfig
	AWSTemplatesDir string
	Boot            bool
}

// NewCmdCreateVault  creates a command object for the "create" command
func NewCmdCreateVault(commonOpts *opts.CommonOptions) *cobra.Command {
	options := &CreateVaultOptions{
		CreateOptions: CreateOptions{
			CommonOptions: commonOpts,
		},
		IngressConfig: kube.IngressConfig{},
	}

	cmd := &cobra.Command{
		Use:     "vault",
		Short:   "Create a new Vault using the vault-operator",
		Long:    createVaultLong,
		Example: createVaultExample,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			helper.CheckErr(err)
		},
	}

	// GKE flags
	cmd.Flags().StringVarP(&options.GKEProjectID, "gke-project-id", "", "", "Google Project ID to use for Vault backend")
	cmd.Flags().StringVarP(&options.GKEZone, "gke-zone", "", "", "The zone (e.g. us-central1-a) where Vault will store the encrypted data")

	awsCreateVaultOptions(cmd, &options.AWSConfig)

	cmd.Flags().StringVarP(&options.ClusterName, "cluster-name", "", "", "Name of the cluster to install vault")
	cmd.Flags().StringVarP(&options.Namespace, "namespace", "n", "", "Namespace where the Vault is created")
	cmd.Flags().StringVarP(&options.SecretsPathPrefix, "secrets-path-prefix", "p", vault.DefaultSecretsPathPrefix, "Path prefix for secrets used for access control config")
	cmd.Flags().BoolVarP(&options.RecreateVaultBucket, "recreate", "", true, "If the bucket already exists delete it so its created empty for the vault")
	cmd.Flags().BoolVarP(&options.NoExposeVault, "no-expose", "", false, "If enabled disable the exposing of the vault")
	cmd.Flags().StringVarP(&options.BucketName, "bucket-name", "", "", "Specify the bucket name. If empty then the bucket name will be based on the vault name")
	cmd.Flags().StringVarP(&options.KeyringName, "keyring-name", "", "", "Specify the KMS Keyring name. If empty then the keyring name will be based on the vault name")
	cmd.Flags().StringVarP(&options.KeyName, "key-name", "", "", "Specify the KMS Key name. If empty then the key name will be based on the vault & keyring name")
	cmd.Flags().StringVarP(&options.ServiceAccountName, "service-account-name", "", "", "Specify Service Account name used. If empty then the service account name will be based on the vault name")

	return cmd
}

func awsCreateVaultOptions(cmd *cobra.Command, options *kubevault.AWSConfig) {
	// AWS flags
	cmd.Flags().BoolVarP(&options.AutoCreate, "aws-auto-create", "", false, "Whether to skip creating resource prerequisites automatically")
	cmd.Flags().StringVarP(&options.DynamoDBRegion, "aws-dynamodb-region", "", "", "The region to use for storing values in AWS DynamoDB")
	cmd.Flags().StringVarP(&options.DynamoDBTable, "aws-dynamodb-table", "", "vault-data", "The table in AWS DynamoDB to use for storing values")
	cmd.Flags().StringVarP(&options.KMSRegion, "aws-kms-region", "", "", "The region of the AWS KMS key to encrypt values")
	cmd.Flags().StringVarP(&options.KMSKeyID, "aws-kms-key-id", "", "", "The ID or ARN of the AWS KMS key to encrypt values")
	cmd.Flags().StringVarP(&options.S3Bucket, "aws-s3-bucket", "", "", "The name of the AWS S3 bucket to store values in")
	cmd.Flags().StringVarP(&options.S3Prefix, "aws-s3-prefix", "", "vault-operator", "The prefix to use for storing values in AWS S3")
	cmd.Flags().StringVarP(&options.S3Region, "aws-s3-region", "", "", "The region to use for storing values in AWS S3")
	cmd.Flags().StringVarP(&options.AccessKeyID, "aws-access-key-id", "", "", "Access key id of service account to be used by vault")
	cmd.Flags().StringVarP(&options.SecretAccessKey, "aws-secret-access-key", "", "", "Secret access key of service account to be used by vault")
}

// Run implements the command
func (o *CreateVaultOptions) Run() error {
	var vaultName string
	if len(o.Args) == 1 {
		vaultName = o.Args[0]
	} else if o.BatchMode {
		return fmt.Errorf("Missing vault name")
	} else {
		// Prompt the user for the vault name
		vaultName, _ = util.PickValue(
			"Vault name:", "", true,
			"The name of the vault that will be created", o.GetIOFileHandles())
	}
	teamSettings, err := o.TeamSettings()
	if err != nil {
		return errors.Wrap(err, "retrieving the team settings")
	}

	if teamSettings.KubeProvider != cloud.GKE && teamSettings.KubeProvider != cloud.AWS && teamSettings.KubeProvider != cloud.EKS {
		return errors.Wrapf(err, "this command only supports the '%s' kubernetes provider", cloud.GKE)
	}

	kubeClient, team, err := o.KubeClientAndNamespace()
	if err != nil {
		return errors.Wrap(err, "creating kubernetes client")
	}

	if o.Namespace == "" {
		o.Namespace = team
	}

	err = kube.EnsureNamespaceCreated(kubeClient, o.Namespace, nil, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to ensure that provided namespace '%s' is created", o.Namespace)
	}

	vaultOperatorClient, err := o.VaultOperatorClient()
	if err != nil {
		return errors.Wrap(err, "creating vault operator client")
	}

	return o.CreateVault(vaultOperatorClient, vaultName, teamSettings.KubeProvider)
}

// CreateVault creates a vault in the existing namespace.
// If the vault already exists, it will error
func (o *CreateVaultOptions) CreateVault(vaultOperatorClient versioned.Interface, vaultName string, kubeProvider string) error {
	// Checks if the vault already exists
	found := kubevault.FindVault(vaultOperatorClient, vaultName, o.Namespace)
	if found {
		return fmt.Errorf("Vault with name '%s' already exists in namespace '%s'", vaultName, o.Namespace)
	}

	kubeClient, _, err := o.KubeClientAndNamespace()
	if err != nil {
		return err
	}

	var clusterName string
	if o.ClusterName == "" {
		clusterName, err = cluster.ShortName(o.Kube())
		if err != nil {
			return err
		}
	} else {
		clusterName = o.ClusterName
	}

	if clusterName == "" {
		return errors.Wrap(err, "unable to determine the cluster name")
	}

	log.Logger().Debugf("cluster short name for vault naming: '%s'", util.ColorInfo(clusterName))

	vaultAuthServiceAccount, err := CreateAuthServiceAccount(kubeClient, vaultName, o.ServiceAccountName, o.Namespace)
	if err != nil {
		return errors.Wrap(err, "creating Vault authentication service account")
	}
	log.Logger().Debugf("Created service account '%s' for Vault authentication", util.ColorInfo(vaultAuthServiceAccount))
	if kubeProvider == cloud.GKE {
		err = o.createVaultGKE(vaultOperatorClient, vaultName, o.BucketName, o.KeyringName, o.KeyName, kubeClient, clusterName, vaultAuthServiceAccount)
	}
	if kubeProvider == cloud.AWS || kubeProvider == cloud.EKS {
		err = o.createVaultAWS(vaultOperatorClient, vaultName, kubeClient, clusterName, vaultAuthServiceAccount)
	}
	if err != nil {
		return errors.Wrap(err, "creating vault")
	}

	// wait for vault service to become ready before finishing the provisioning
	err = services.WaitForService(kubeClient, vaultName, o.Namespace, 1*time.Minute)
	if err != nil {
		return errors.Wrap(err, "waiting for vault service")
	}

	if o.NoExposeVault {
		log.Logger().Infof("not exposing vault '%s' exposed", vaultName)
		return nil
	}
	log.Logger().Infof("Exposing Vault...")
	err = o.exposeVault(vaultName)
	if err != nil {
		return errors.Wrap(err, "exposing vault")
	}
	log.Logger().Infof("Vault '%s' exposed", util.ColorInfo(vaultName))
	return nil
}

func (o *CreateVaultOptions) dockerImages() (map[string]string, error) {
	images := map[string]string{
		kubevault.BankVaultsImage: "",
		kubevault.VaultImage:      "",
	}

	resolver, err := o.CreateVersionResolver("", "")
	if err != nil {
		return images, errors.Wrap(err, "creating the docker image version resolver")
	}
	for image := range images {
		version, err := resolver.ResolveDockerImage(image)
		if err != nil {
			return images, errors.Wrapf(err, "resolving docker image %q", image)
		}
		images[image] = version
	}
	return images, nil
}

func (o *CreateVaultOptions) createVaultGKE(vaultOperatorClient versioned.Interface, vaultName string, bucketName string,
	keyringName string, keyName string, kubeClient kubernetes.Interface, clusterName string, vaultAuthServiceAccount string) error {
	err := o.GCloud().Login("", true)
	if err != nil {
		return errors.Wrap(err, "login into GCP")
	}

	if o.GKEProjectID == "" {
		if kubeClient, ns, err := o.KubeClientAndDevNamespace(); err == nil {
			if data, err := kube.ReadInstallValues(kubeClient, ns); err == nil && data != nil {
				o.GKEProjectID = data[kube.ProjectID]
				if o.GKEZone == "" {
					o.GKEZone = data[kube.Zone]
				}
			}
		}
	}

	if o.GKEProjectID == "" {
		o.GKEProjectID, err = o.GetGoogleProjectID("")
		if err != nil {
			return err
		}
	}

	err = o.CreateOptions.CommonOptions.RunCommandVerbose(
		"gcloud", "config", "set", "project", o.GKEProjectID)
	if err != nil {
		return err
	}

	if o.GKEZone == "" {
		defaultZone := ""
		if cluster, err := cluster.Name(o.Kube()); err == nil && cluster != "" {
			if clusterZone, err := o.GCloud().ClusterZone(cluster); err == nil {
				defaultZone = clusterZone
			}
		}

		zone, err := o.GetGoogleZoneWithDefault(o.GKEProjectID, defaultZone)
		if err != nil {
			return err
		}
		o.GKEZone = zone
	}

	log.Logger().Debugf("Ensure KMS API is enabled")
	err = o.GCloud().EnableAPIs(o.GKEProjectID, "cloudkms")
	if err != nil {
		return errors.Wrap(err, "unable to enable 'cloudkms' API")
	}

	log.Logger().Debugf("Creating GCP service account for Vault backend")
	gcpServiceAccountSecretName, err := gkevault.CreateVaultGCPServiceAccount(o.GCloud(), kubeClient, vaultName, o.Namespace, clusterName, o.GKEProjectID)
	if err != nil {
		return errors.Wrap(err, "creating GCP service account")
	}
	log.Logger().Debugf("'%s' service account created", util.ColorInfo(gcpServiceAccountSecretName))

	log.Logger().Debugf("Setting up GCP KMS configuration")
	kmsConfig, err := gkevault.CreateKmsConfig(o.GCloud(), vaultName, keyringName, keyName, o.GKEProjectID)
	if err != nil {
		return errors.Wrap(err, "creating KMS configuration")
	}
	log.Logger().Debugf("KMS Key '%s' created in keying '%s'", util.ColorInfo(kmsConfig.Key), util.ColorInfo(kmsConfig.Keyring))

	vaultBucket, err := gkevault.CreateBucket(o.GCloud(), vaultName, bucketName, o.GKEProjectID, o.GKEZone, o.RecreateVaultBucket, o.BatchMode, o.GetIOFileHandles())
	if err != nil {
		return errors.Wrap(err, "creating Vault GCS data bucket")
	}
	log.Logger().Infof("GCS bucket '%s' was created for Vault backend", util.ColorInfo(vaultBucket))

	log.Logger().Infof("Creating Vault...")
	gcpConfig := &kubevault.GCPConfig{
		ProjectId:   o.GKEProjectID,
		KmsKeyring:  kmsConfig.Keyring,
		KmsKey:      kmsConfig.Key,
		KmsLocation: kmsConfig.Location,
		GcsBucket:   vaultBucket,
	}
	images, err := o.dockerImages()
	if err != nil {
		return errors.Wrap(err, "loading docker images from versions repository")
	}
	err = kubevault.CreateGKEVault(kubeClient, vaultOperatorClient, vaultName, o.Namespace, images, gcpServiceAccountSecretName,
		gcpConfig, vaultAuthServiceAccount, o.Namespace, o.SecretsPathPrefix)
	if err != nil {
		return errors.Wrap(err, "creating vault")
	}
	log.Logger().Infof("Vault '%s' created in cluster '%s'", util.ColorInfo(vaultName), util.ColorInfo(clusterName))
	return nil
}

func (o *CreateVaultOptions) createVaultAWS(vaultOperatorClient versioned.Interface, vaultName string,
	kubeClient kubernetes.Interface, clusterName string, vaultAuthServiceAccount string) error {

	if o.AutoCreate {
		_, clusterRegion, err := amazon.GetCurrentlyConnectedRegionAndClusterName()
		if err != nil {
			return errors.Wrap(err, "finding default AWS region")
		}

		if err := o.ApplyDefaultRegionIfEmpty(clusterRegion); err != nil {
			return errors.Wrap(err, "setting the default region")
		}

		domain := "jenkins-x-domain"
		username := o.ProvidedIAMUsername
		if username == "" {
			username = "vault_" + clusterRegion
		}

		bucketName := o.S3Bucket
		if bucketName == "" {
			bucketName = "vault-unseal." + o.S3Region + "." + domain
		}

		valueUUID, err := uuid.NewV4()
		if err != nil {
			return errors.Wrapf(err, "Generating UUID failed")
		}

		// Create suffix to apply to resources
		suffixString := valueUUID.String()[:7]
		var kmsID, s3Name, tableName, accessID, secretKey *string
		if o.Boot {
			accessID, secretKey, kmsID, s3Name, tableName, err = awsvault.CreateVaultResourcesBoot(awsvault.ResourceCreationOpts{
				Region:          clusterRegion,
				Domain:          domain,
				Username:        username,
				BucketName:      bucketName,
				TableName:       autoCreateTableName,
				AWSTemplatesDir: o.AWSTemplatesDir,
				UniqueSuffix:    suffixString,
			})
		} else {
			// left for non-boot clusters until deprecation
			accessID, secretKey, kmsID, s3Name, tableName, err = awsvault.CreateVaultResources(awsvault.ResourceCreationOpts{
				Region:     clusterRegion,
				Domain:     domain,
				Username:   username,
				BucketName: bucketName,
				TableName:  autoCreateTableName,
			})
		}

		if err != nil {
			return errors.Wrap(err, "an error occurred while creating the vault resources")
		}
		if s3Name != nil {
			o.S3Bucket = *s3Name
		}
		if kmsID != nil {
			o.KMSKeyID = *kmsID
		}
		if tableName != nil {
			o.DynamoDBTable = *tableName
		}
		if accessID != nil {
			o.AccessKeyID = *accessID
		}
		if secretKey != nil {
			o.SecretAccessKey = *secretKey
		}

	} else {
		if o.S3Bucket == "" {
			return fmt.Errorf("missing S3 bucket flag")
		}
		if o.KMSKeyID == "" {
			return fmt.Errorf("missing AWS KMS key id flag")
		}
		if o.AccessKeyID == "" {
			return fmt.Errorf("missing AWS access key id flag")
		}
		if o.SecretAccessKey == "" {
			return fmt.Errorf("missing AWS secret access key flag")
		}

		if err := o.ApplyDefaultRegionIfEmpty(""); err != nil {
			return errors.Wrap(err, "setting the default region")
		}
	}

	awsServiceAccountSecretName, err := awsvault.StoreAWSCredentialsIntoSecret(kubeClient, o.AccessKeyID, o.SecretAccessKey, vaultName, o.Namespace)
	if err != nil {
		return errors.Wrap(err, "storing the service account credentials into a secret")
	}
	images, err := o.dockerImages()
	if err != nil {
		return errors.Wrap(err, "loading docker images from versions repository")
	}
	err = kubevault.CreateAWSVault(kubeClient, vaultOperatorClient, vaultName, o.Namespace, images,
		awsServiceAccountSecretName, &o.AWSConfig, vaultAuthServiceAccount, o.Namespace, o.SecretsPathPrefix)
	if err != nil {
		return errors.Wrap(err, "creating vault")
	}
	log.Logger().Infof("Vault '%s' created in cluster '%s'", util.ColorInfo(vaultName), util.ColorInfo(clusterName))

	return nil
}

func (o *CreateVaultOptions) exposeVault(vaultService string) error {
	client, err := o.KubeClient()
	if err != nil {
		return err
	}
	svc, err := client.CoreV1().Services(o.Namespace).Get(vaultService, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "getting the vault service: %s", vaultService)
	}
	if svc.Annotations == nil {
		svc.Annotations = map[string]string{}
	}
	if svc.Annotations[kube.AnnotationExpose] == "" {
		svc.Annotations[kube.AnnotationExpose] = "true"
		svc.Annotations[kube.AnnotationExposePort] = exposedVaultPort
		svc, err = client.CoreV1().Services(o.Namespace).Update(svc)
		if err != nil {
			return errors.Wrapf(err, "updating %s service annotations", vaultService)
		}
	}

	upgradeIngOpts := &upgrade.UpgradeIngressOptions{
		CommonOptions:       o.CommonOptions,
		Namespaces:          []string{o.Namespace},
		Services:            []string{vaultService},
		IngressConfig:       o.IngressConfig,
		SkipResourcesUpdate: true,
		WaitForCerts:        true,
	}
	return upgradeIngOpts.Run()
}

// CreateAuthServiceAccount creates a Serivce Account for the Auth service for vault
func CreateAuthServiceAccount(client kubernetes.Interface, vaultName, serviceAccountName string, namespace string) (string, error) {
	if serviceAccountName == "" {
		serviceAccountName = AuthServiceAccountName(vaultName)
	}

	_, err := serviceaccount.CreateServiceAccount(client, namespace, serviceAccountName)
	if err != nil {
		return "", errors.Wrap(err, "creating vault auth service account")
	}
	return serviceAccountName, nil
}

// AuthServiceAccountName creates a service account name for a given vault and cluster name
func AuthServiceAccountName(vaultName string) string {
	return fmt.Sprintf("%s-%s", vaultName, "auth-sa")
}

// ApplyDefaultRegionIfEmpty applies the default region to all AWS resources
func (o *CreateVaultOptions) ApplyDefaultRegionIfEmpty(enforcedDefault string) error {
	if o.DynamoDBRegion == "" || o.KMSRegion == "" || o.S3Region == "" {
		var defaultRegion string
		var err error
		if enforcedDefault == "" {
			_, defaultRegion, err = amazon.GetCurrentlyConnectedRegionAndClusterName()
			if err != nil {
				return errors.Wrap(err, "finding default AWS region")
			}
		} else {
			defaultRegion = enforcedDefault
		}

		log.Logger().Infof("Region not specified, defaulting to %s", util.ColorInfo(defaultRegion))
		if o.DynamoDBRegion == "" {
			o.DynamoDBRegion = defaultRegion
		}
		if o.KMSRegion == "" {
			o.KMSRegion = defaultRegion
		}
		if o.S3Region == "" {
			o.S3Region = defaultRegion
		}

	}
	return nil
}
