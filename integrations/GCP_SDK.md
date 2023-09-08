# GCP SDK

In order to use Sidekick with your gcp sdk, you need to update the GCS client endpoint to point to the sidekick url (ex: `localhost:7075`).

Here are some examples of how to use various gcp sdks to work with sidekick:

1. [gsutil](#gsutil)
2. [go](#go)
3. [nodejs](#nodejs)

<a name="gsutil"></a>

## gsutil

```bash
gsutil  -o Credentials:gs_json_host=127.0.0.1 -o "Credentials:gs_json_port=7076" -o "Boto:https_validate_certificates=False" ls -r gs://<YOUR_BUCKET>
```

```bash
gsutil  -o Credentials:gs_json_host=127.0.0.1 -o "Credentials:gs_json_port=7076" -o "Boto:https_validate_certificates=False" cp gs://<YOUR_BUCKET>/<YOUR_OBJECT_KEY> <LOCAL_FILE_NAME>
```

```bash
gsutil  -o Credentials:gs_json_host=127.0.0.1 -o "Credentials:gs_json_port=7076" -o "Boto:https_validate_certificates=False" cp <LOCAL_FILE_NAME> gs://<YOUR_BUCKET>/<YOUR_OBJECT_KEY>
```

<a name="go"></a>

## Go

```go
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithEndpoint("http://127.0.0.1:7075"))
	if err != nil {
		log.Fatal(err)
	}
	err = listObjects(ctx, client, "km-integration-tests-sep5-0")
	if err != nil {
		log.Printf("listObjects failed: %v", err)
		log.Fatal(err)
	}
	err = getObject(ctx, client, "km-integration-tests-sep5-0", "animals/1.csv")
	if err != nil {
		log.Printf("getObject failed: %v", err)
		log.Fatal(err)
	}
	err = putObject(ctx, client, "km-integration-tests-sep5-0", "kote/kote.txt", "kote.txt")
	if err != nil {
		log.Printf("putObject failed: %v", err)
		log.Fatal(err)
	}
}

func listObjects(ctx context.Context, client *storage.Client, bucket string) error {
	it := client.Bucket(bucket).Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		fmt.Println(attrs.Name)
	}
	return nil
}

func getObject(ctx context.Context, client *storage.Client, bucket string, key string) error {
	r, err := client.Bucket(bucket).Object(key).NewReader(ctx)
	if err != nil {
		return err
	}
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	fmt.Println(string(buf))
	return nil
}

func putObject(ctx context.Context, client *storage.Client, bucket string, key string, filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	w := client.Bucket(bucket).Object(key).NewWriter(ctx)
	if _, err = io.Copy(w, f); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}
	fmt.Printf("File at %v uploaded to %s/%s\n", filepath, bucket, key)
	return nil
}

```

<a name="nodejs"></a>

## Node.js

```js
TODO;
```
