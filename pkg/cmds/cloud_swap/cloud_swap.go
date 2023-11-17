/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloud_swap

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/spf13/cobra"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	"gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
	"gomodules.xyz/blobfs"
)

var (
	srcBucketURL       string
	dstBucketURL       string
	localBackupDir     string
	disableLocalBackup bool

	ms struct {
		bucket    string
		endpoint  string
		accessID  string
		secretKey string
	}

	mc                    *minio.Client
	srcBucket             *blob.Bucket
	srcFS, dstFS, localFS blobfs.Interface
)

func NewCmdCloudSwap() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud-swap",
		Short: "Copy data from one cloud storage to another",
		Example: `
# Export required ENV for the buckets
export GOOGLE_APPLICATION_CREDENTIALS=<google-sa-cred-path>
export AWS_ACCESS_KEY_ID=<s3-bucket-access-key>
export AWS_SECRET_ACCESS_KEY=<s3-bucket-secret-access-key>

# Now copy the files
ace cloud-swap --src-bucket-url="gs://<google-bucket-name>" \
    --dst-bucket-url="s3://<s3-compatible-bucket>?region=<us-east-1>&endpoint=<bucket-endpoint>"
`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return copyDataToDestination(cmd.Context())
		},
	}

	cmd.Flags().StringVar(&ms.endpoint, "minio.endpoint", "", "MinIO storage endpoint")
	cmd.Flags().StringVar(&ms.bucket, "minio.bucket", "", "MinIO storage bucket name")
	cmd.Flags().StringVar(&ms.accessID, "minio.access-id", "", "ACCESS_KEY_ID for MinIO storage")
	cmd.Flags().StringVar(&ms.secretKey, "minio.secret-key", "", "SECRET_ACCESS_KEY for MinIO storage")

	cmd.Flags().StringVar(&srcBucketURL, "src-bucket-url", "", "Complete source-bucket url with scheme, region, endpoints")
	cmd.Flags().StringVar(&dstBucketURL, "dst-bucket-url", "", "Complete destination-bucket url with scheme, region, endpoints")
	cmd.Flags().StringVar(&localBackupDir, "local-backup-dir", localDefaultDir(), "Temporary local backup")
	cmd.Flags().BoolVar(&disableLocalBackup, "disable-local-backup", false, "Disable local backup")
	if err := cmd.MarkFlagRequired("dst-bucket-url"); err != nil {
		log.Fatal(err)
	}

	cmd.MarkFlagsMutuallyExclusive("minio.endpoint", "src-bucket-url")
	cmd.MarkFlagsRequiredTogether("minio.endpoint", "minio.bucket", "minio.access-id", "minio.secret-key")

	return cmd
}

func copyDataToDestination(ctx context.Context) error {
	err := initiateBuckets(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Copying files to destination storage ...\n\n")

	if mc != nil {
		if err = copyFromMinio(ctx); err != nil {
			return err
		}
	} else {
		if err = copyFromSrc(ctx); err != nil {
			return err
		}
	}

	fmt.Println("\nFile copying completed successfully!")
	fmt.Printf("A local copy can be found in `%s` directory\n", strings.TrimPrefix(localBackupDir, fmt.Sprintf("%s://", fileblob.Scheme)))
	return nil
}
