package restapi

import (
	"github.com/Mastercard/terraform-provider-restapi/fakeserver"
	mylog "github.com/Mastercard/terraform-provider-restapi/log"
	"github.com/hashicorp/terraform/helper/resource"
	"os"
	"testing"
)

func TestAccRestApiObject_importBasic(t *testing.T) {
	log := mylog.New(debug)
	apiServerObjects := make(map[string]map[string]interface{})

	svr := fakeserver.NewFakeServer(&fakeserver.Opts{
		Port:    8082,
		Objects: apiServerObjects,
		Start:   true,
		Debug:   debug,
		Logger:  log,
		Dir:     "",
	})
	os.Setenv("REST_API_URI", "http://127.0.0.1:8082")

	resource.UnitTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { svr.StartInBackground() },
		Steps: []resource.TestStep{
			{
				Config: generate_test_resource(
					"Foo",
					`{ "id": "1234", "first": "Foo", "last": "Bar" }`,
					make(map[string]interface{}),
				),
			},
			{
				ResourceName:            "restapi_object.Foo",
				ImportState:             true,
				ImportStateId:           "1234",
				ImportStateIdPrefix:     "/api/objects/",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"debug", "data"},
			},
		},
	})

	svr.Shutdown()
}
