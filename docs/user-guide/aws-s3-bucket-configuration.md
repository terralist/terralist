# AWS S3 Bucket Configuration

Terralist supports storing artifacts on S3-compatible backends. This resource presents how you can configure the bucket on AWS S3, so if you are looking for another backend, refer to the `s3-` prefixed [configuration options](../configuration.md).

!!! note "The following examples will contain placeholders in the form of `{placeholder}`. You are expected to replace those placeholders with values that suits your needs."

## Access

For Terralist to be able to access the S3 bucket, you must create an identity for it. There are two possible options:

- AWS IAM Role (**recommended**)
- AWS IAM User

The bucket can either be created using the legacy ACL system, or with the newer, recommended, bucket policy system. Also, the bucket can either be in the same AWS account with the IAM identity or not. Depending on how and where the bucket is configured, you will either have to attach a policy to the IAM identity or not. Possible cases:

| Same Account | ACLs or Policy | Should attach IAM identity policy |
| ------------ | -------------- | --------------------------------- |
| Yes          | ACLs           | Optional                          |
| Yes          | Policy         | Optional                          |
| No           | ACLs           | Yes                               |
| No           | Policy         | Yes                               |

If you happen to be in the case where you are not required to attach the IAM identity, you may skip to the [Bucket Policy](#bucket-policy) section, as the following policy becomes optional.

```json title="policy.json"
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "FindBucket",
      "Effect": "Allow",
      "Action": "s3:ListBucket",
      "Resource": "arn:aws:s3:::{s3-bucket-name}"
    },
    {
      "Sid": "UseBucket",
      "Effect": "Allow",
       "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject"
      ],
      "Resource": "arn:aws:s3:::{s3-bucket-name}/{s3-bucket-prefix}/*"
    },
    {
      "Sid": "DecryptSSE",
      "Effect": "Allow",
      "Action": [
        "kms:Decrypt",
        "kms:DescribeKey"
      ],
      "Resource": "arn:aws:kms:{REGION}:{AWS-ACCOUNT-ID}:key/{KEY-ID}",
    }
  ]
}
```

!!! note "Notice that if you don't want to set a given bucket prefix within your bucket, the second statement (sid = `UseBucket`) should have the resource set to `arn:aws:s3:::{s3-bucket-name}/*`."

!!! note "The third statement (sid = `DecryptSSE`) is required only if the `s3-server-side-encryption` configuration option is set."

## Bucket Policy

To grant Terralist access to the bucket, the following policy should be applied as bucket policy.

```json title="policy.json"
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "UseBucket",
      "Effect": "Allow",
       "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject"
      ],
      "Resource": "arn:aws:s3:::{s3-bucket-name}/{s3-bucket-prefix}/*",
      "Principal": {
        "AWS": "arn:aws:iam::{AWS-ACCOUNT-ID}:{user/role}/{identity-name}"
      },
    }
  ]
}
```

!!! note "Notice that if you don't want to set a given bucket prefix within your bucket, the resource should be set to `arn:aws:s3:::{s3-bucket-name}/*`."
