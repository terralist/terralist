package file

import "testing"

func TestCheckSource(t *testing.T) {

	for _, src := range []string{
		"https://example.com/module.zip",
		"http://example.com/module.tar.gz",
		"git::https://github.com/hashicorp/terraform-template-dir?ref=v1.0.2",
		"git::https://github.com/terraform-aws-modules/terraform-aws-vpc//modules/vpc-endpoints?ref=v5.0.0",
		"git::ssh://git@github.com/hashicorp/terraform-template-dir.git",
		"github.com/hashicorp/terraform-template-dir",
	} {
		t.Run("accepts "+src, func(t *testing.T) {
			if err := checkSource(src, true); err != nil {
				t.Fatalf("expected %q to be accepted, got %v", src, err)
			}
		})
	}

	for _, src := range []string{
		"s3::https://s3.amazonaws.com/bucket/module.zip",
		"bucket.s3.amazonaws.com/module.zip",
		"gcs::https://www.googleapis.com/storage/v1/bucket/module.zip",
		"hg::https://example.com/repo",
		"git::file:///etc/terralist",
		"file:///etc/terralist",
	} {
		t.Run("refuses "+src, func(t *testing.T) {
			if err := checkSource(src, true); err == nil {
				t.Fatalf("expected %q to be refused", src)
			}
		})
	}

	for _, src := range []string{
		"git::https://10.0.0.5/repo.git",
		"git::ssh://git@127.0.0.1/repo.git",
		"git::https://localhost/repo.git",
	} {
		t.Run("refuses the private "+src, func(t *testing.T) {
			if err := checkSource(src, false); err == nil {
				t.Fatalf("expected %q to be refused", src)
			}
		})

		t.Run("accepts the private "+src+" when private addresses are allowed", func(t *testing.T) {
			if err := checkSource(src, true); err != nil {
				t.Fatalf("expected %q to be accepted, got %v", src, err)
			}
		})
	}
}
