package vault

import (
	"fmt"

	"github.com/jenkins-x/jx/pkg/cloud/gke"
	"github.com/jenkins-x/jx/pkg/log"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
)

const (
	gkeServiceAccountSecretKey = "service-account.json"
	//DefaultVaultAbbreviation is vault service accounts suffix
	DefaultVaultAbbreviation = "vt"
)

var (
	ServiceAccountRoles = []string{"roles/storage.objectAdmin",
		"roles/cloudkms.admin",
		"roles/cloudkms.cryptoKeyEncrypterDecrypter",
	}
)

// KmsConfig keeps the configuration for Google KMS service
type KmsConfig struct {
	Keyring  string
	Key      string
	Location string
	project  string
}

// This is a loose collection of methods needed to set up a vault in GKE.
// If they are generic enough and needed elsewhere, we can move them up one level to more generic GCP methods.

// CreateKmsConfig creates a KMS config for the GKE Vault
func CreateKmsConfig(gcloud gke.GClouder, vaultName, keyringName string, keyName string, projectID string) (*KmsConfig, error) {
	if keyringName == "" {
		keyringName = gke.KeyringName(vaultName)
	}
	config := &KmsConfig{
		Keyring:  keyringName,
		Key:      keyName,
		Location: gke.KmsLocation,
		project:  projectID,
	}

	err := gcloud.CreateKmsKeyring(config.Keyring, config.project)
	if err != nil {
		return nil, errors.Wrapf(err, "creating kms keyring '%s'", config.Keyring)
	}

	if config.Key == "" {
		config.Key = gke.KeyName(vaultName)
	}

	err = gcloud.CreateKmsKey(config.Key, config.Keyring, config.project)
	if err != nil {
		return nil, errors.Wrapf(err, "creating the kms key '%s'", config.Key)
	}
	return config, nil
}

// CreateGCPServiceAccount creates a service account in GCP for the vault service
func CreateVaultGCPServiceAccount(gcloud gke.GClouder, kubeClient kubernetes.Interface, vaultName, namespace, clusterName, projectID string) (string, error) {

	gcpServiceAccountSecretName, error := gcloud.CreateGCPServiceAccount(kubeClient, vaultName, DefaultVaultAbbreviation, namespace, clusterName, projectID, ServiceAccountRoles, gkeServiceAccountSecretKey)

	if error != nil {
		return "", errors.Wrap(error, "creating the Vault GCP Service Account")
	}
	return gcpServiceAccountSecretName, nil
}

// CreateBucket Creates a bucket in GKE to store the backend (encrypted) data for vault
func CreateBucket(gcloud gke.GClouder, vaultName, bucketName string, projectID, zone string, recreate bool, batchMode bool, handles util.IOFileHandles) (string, error) {
	if bucketName == "" {
		bucketName = gke.BucketName(vaultName)
	}
	exists, err := gcloud.BucketExists(projectID, bucketName)
	if err != nil {
		return "", errors.Wrap(err, "checking if Vault GCS bucket exists")
	}
	if exists {
		if !recreate {
			return bucketName, nil
		}
		if batchMode {
			log.Logger().Warnf("We are deleting the Vault bucket %s so that Vault will install cleanly", bucketName)
		} else {
			if !util.Confirm(fmt.Sprintf("We are about to delete bucket %q, so we can install a clean Vault. Are you sure: ", bucketName),
				true, "We recommend you delete the Vault bucket on install to ensure Vault starts up reliably", handles) {
				return bucketName, nil
			}
		}
		err = gcloud.DeleteAllObjectsInBucket(bucketName)
		if err != nil {
			return "", errors.Wrapf(err, "failed to remove objects from GCS bucket %s", bucketName)
		}
	}

	if zone == "" {
		return "", errors.New("GKE zone must be provided in 'gke-zone' option")
	}
	region := gke.GetRegionFromZone(zone)
	err = gcloud.CreateBucket(projectID, bucketName, region)
	if err != nil {
		return "", errors.Wrap(err, "creating Vault GCS bucket")
	}
	return bucketName, nil
}
