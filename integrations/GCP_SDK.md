# GCP SDK

In order to use Sidekick with your gcp sdk, you need to update the GCS client endpoint to point to the sidekick url (ex: `localhost:7075`).

Here are some examples of how to use various gcp sdks to work with sidekick:

1. [gsutil](#gsutil)

<a name="aws-cli"></a>

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
