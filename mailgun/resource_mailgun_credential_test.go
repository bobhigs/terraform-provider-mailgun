package mailgun

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mailgun/mailgun-go/v3"
)

func TestAccMailgunDomainCredential_Basic(t *testing.T) {
	domain := os.Getenv("MAILGUN_TEST_DOMAIN")

	if domain == "" {
		t.Fatal("MAILGUN_TEST_DOMAIN must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: newProvider(),
		CheckDestroy:      testAccCheckMailgunCrendentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMailgunCredentialConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMailgunCredentialExists("mailgun_domain_credential.foobar"),
					resource.TestCheckResourceAttr(
						"mailgun_domain_credential.foobar", "domain", domain),
					resource.TestCheckResourceAttr(
						"mailgun_domain_credential.foobar", "email", "test_crendential@"+domain),
					resource.TestCheckResourceAttr(
						"mailgun_domain_credential.foobar", "password", "supersecretpassword1234"),
					resource.TestCheckResourceAttr(
						"mailgun_domain_credential.foobar", "region", "us"),
				),
			},
		},
	})
}

func testAccCheckMailgunCrendentialDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mailgun_domain_credential" {
			continue
		}

		itCredentials := client.MailgunClient.ListCredentials(nil)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		var page []mailgun.Credential

		for itCredentials.Next(ctx, &page) {

			for _, c := range page {
				if c.Login == rs.Primary.ID {
					return fmt.Errorf("The credential '%s' found! Created at: %s", rs.Primary.ID, c.CreatedAt.String())
				}
			}
		}

		if err := itCredentials.Err(); err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckMailgunCredentialExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No domain credential ID is set")
		}

		client := testAccProvider.Meta().(*Config)
		itCredentials := client.MailgunClient.ListCredentials(nil)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		var page []mailgun.Credential

		for itCredentials.Next(ctx, &page) {
			for _, c := range page {
				if c.Login == rs.Primary.ID {
					return nil
				}
			}
		}

		if err := itCredentials.Err(); err != nil {
			return err
		}

		return fmt.Errorf("The credential '%s' not found!", rs.Primary.ID)
	}
}

func testAccCheckMailgunCredentialConfig(domain string) string {
	return `resource "mailgun_domain_credential" "foobar" {
	domain = "` + domain + `"
	email = "test_crendential@` + domain + `"
	password = "supersecretpassword1234"
	region = "us"
}`
}