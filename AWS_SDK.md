# AWS SDK

In order to use sidekick with your aws sdk, you need to update the S3 Client hostname to point to the sidekick url (ex: `localhost:7075`).
Currently you also need to set your s3 client to use `pathStyle` to work.

Here are some examples of how to use various aws sdks to work with sidekick:

1. [Aws cli](#aws-cli)
1. [Go](#go)
1. [Java](#java)

<a name="aws-cli"></a>
## AWS cli

```bash
aws s3api get-object --bucket <YOUR_BUCKET> --key <YOUR_OBJECT_KEY>  delete_me.csv --endpoint-url http://localhost:7075
```

<a name="go"></a>
## Go

```Go
package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	ctx := context.Background()
	sidekickURL := "http://localhost:7075"
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           sidekickURL,
				SigningRegion: region,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})
	cfg, _ := config.LoadDefaultConfig(ctx, config.WithEndpointResolverWithOptions(customResolver))
	s3c := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	awsBucket := aws.String("MY_BUCKET")
	awsKey := aws.String("foo.txt")

	// PutObject
	reader := strings.NewReader("Hello World!")
	if _, err := s3c.PutObject(ctx, &s3.PutObjectInput{
		Bucket: awsBucket,
		Key:    awsKey,
		Body:   reader,
	}); err != nil {
		panic(err)
	}

	// GetObject
	getObjResp, err := s3c.GetObject(ctx, &s3.GetObjectInput{
		Bucket: awsBucket,
		Key:    awsKey,
	})
	if err != nil {
		panic(err)
	}

	data, err := io.ReadAll(getObjResp.Body)
	getObjResp.Body.Close()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
```
<a name="java"></a>
## Java

Currently the Java sdk by default uses the [streamin signature](https://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-streaming.html) when uploading objects. Sidekick does not currenlty support this and you need to disable the chunk encoding when creating the client as shown below.

``` java
import com.amazonaws.AmazonServiceException;
import com.amazonaws.regions.Regions;
import com.amazonaws.services.s3.AmazonS3;
import com.amazonaws.services.s3.AmazonS3ClientBuilder;
import com.amazonaws.services.s3.model.S3Object;
import com.amazonaws.services.s3.model.S3ObjectInputStream;

import java.io.File;
import java.io.FileNotFoundException;
import java.io.FileOutputStream;
import java.io.IOException;
import java.nio.file.Paths;

/**
 * Upload a file to an Amazon S3 bucket.
 * 
 * This code expects that you have AWS credentials set up per:
 * http://docs.aws.amazon.com/java-sdk/latest/developer-guide/setup-credentials.html
 */
public class PutObject {
    public static void main(String[] args) {
        String bucket_name = "sidekick-test-rvh-west-2";
        String file_path = "foo.txt";
        String key_name = Paths.get(file_path).getFileName().toString();

        System.out.format("Uploading %s to S3 bucket %s...\n", file_path, bucket_name);
        final AmazonS3 s3 = AmazonS3ClientBuilder.standard()
                .withPathStyleAccessEnabled(true)
                .withEndpointConfiguration(new AmazonS3ClientBuilder.EndpointConfiguration("http://localhost:7075",
                        Regions.DEFAULT_REGION.getName()))
                .disableChunkedEncoding()  // This is needed in order for puObject to work
                .build();
        try {
            // PUT OBJECT
            s3.putObject(bucket_name, key_name, new File(file_path));

            // GET OBJECT
            S3Object o = s3.getObject(bucket_name, key_name);
            S3ObjectInputStream s3is = o.getObjectContent();
            FileOutputStream fos = new FileOutputStream(new File(key_name));
            byte[] read_buf = new byte[1024];
            int read_len = 0;
            while ((read_len = s3is.read(read_buf)) > 0) {
                fos.write(read_buf, 0, read_len);
            }
            s3is.close();
            fos.close();
        } catch (AmazonServiceException e) {
            System.err.println(e.getErrorMessage());
            System.exit(1);
        } catch (FileNotFoundException e) {
            System.err.println(e.getMessage());
            System.exit(1);
        } catch (IOException e) {
            System.err.println(e.getMessage());
            System.exit(1);
        }
        System.out.println("Done!");
    }
}
```