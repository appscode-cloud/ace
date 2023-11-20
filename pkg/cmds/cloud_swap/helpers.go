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
	goerr "errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"gocloud.dev/blob"
	"gomodules.xyz/blobfs"
	"k8s.io/client-go/util/homedir"
)

func initiateBuckets(ctx context.Context) (err error) {
	if ms.endpoint != "" {
		mc, err = minio.New(ms.endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(ms.accessID, ms.secretKey, ""),
			Secure: tlsEnabled(ms.endpoint),
		})
		if err != nil {
			return err
		}
	} else {
		srcBucket, err = blob.OpenBucket(ctx, srcBucketURL)
		if err != nil {
			return err
		}
		srcFS = blobfs.New(srcBucketURL)
	}

	dstFS = blobfs.New(dstBucketURL)

	// initiate local backup bucket
	if !disableLocalBackup {
		if _, err = os.Stat(localBackupDir); goerr.Is(err, os.ErrNotExist) {
			if err = os.Mkdir(localBackupDir, os.ModePerm); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		if !strings.HasPrefix(localBackupDir, "file://") {
			localBackupDir = fmt.Sprintf("%s%s", "file://", localBackupDir)
		}

		localFS = blobfs.New(localBackupDir)
	}

	return nil
}

func localDefaultDir() string {
	home := homedir.HomeDir()
	return path.Join(home, "cloud-swap-backup")
}

func tlsEnabled(url string) bool {
	resp, err := http.Get("https://" + url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.TLS != nil
}

func copyFromMinio(ctx context.Context) error {
	fileCount := 0
	for oi := range mc.ListObjects(ctx, ms.bucket, minio.ListObjectsOptions{Recursive: true}) {
		if oi.Err != nil {
			return oi.Err
		}
		fileCount += 1
		obj, err := mc.GetObject(ctx, ms.bucket, oi.Key, minio.GetObjectOptions{})
		if err != nil {
			return err
		}

		data, err := io.ReadAll(obj)
		if err != nil {
			return err
		}
		if err = writeDataToDestination(ctx, oi.Key, data); err != nil {
			return err
		}
		fmt.Printf("%6d. %s\n", fileCount, oi.Key)
	}

	return nil
}

func copyFromSrc(ctx context.Context) (err error) {
	var objs []*blob.ListObject
	fileCount := 0
	pageToken := blob.FirstPageToken
	for pageToken != nil {
		objs, pageToken, err = srcBucket.ListPage(ctx, pageToken, 20, &blob.ListOptions{})
		if err != nil {
			return err
		}

		for _, obj := range objs {
			fileCount += 1
			data, err := srcFS.ReadFile(ctx, obj.Key)
			if err != nil {
				return errors.Wrapf(err, "read file from source")
			}

			if err = writeDataToDestination(ctx, obj.Key, data); err != nil {
				return err
			}
			fmt.Printf("%6d. %s\n", fileCount, obj.Key)
		}
	}
	return nil
}

func writeDataToDestination(ctx context.Context, fileName string, data []byte) error {
	if !disableLocalBackup {
		if err := localFS.WriteFile(ctx, fileName, data); err != nil {
			return errors.Wrapf(err, "write file to local backup")
		}
	}

	if err := dstFS.WriteFile(ctx, fileName, data); err != nil {
		return errors.Wrapf(err, "write file to destination")
	}
	return nil
}
