package restapi

import (
	"github.com/Mastercard/terraform-provider-restapi/fakeserver"
	mylog "github.com/Mastercard/terraform-provider-restapi/log"
	"github.com/hashicorp/terraform/helper/resource"
	"os"
	"testing"
)

func TestAccRestApiObject_importBasic(t *testing.T) {
	debug := false
	api_server_objects := make(map[string]map[string]interface{})

	svr := fakeserver.NewFakeServer(&fakeserver.Opts{
		Port:    8082,
		Objects: api_server_objects,
		Start:   true,
		Debug:   debug,
		Logger:  mylog.New(debug),
		Dir:     "",
	})
	os.Setenv("REST_API_URI", "http://127.0.0.1:8082")

	opt := &apiClientOpt{
		uri:                   "http://127.0.0.1:8082/",
		insecure:              false,
		username:              "",
		password:              "",
		headers:               make(map[string]string, 0),
		timeout:               2,
		id_attribute:          "id",
		copy_keys:             make([]string, 0),
		write_returns_object:  false,
		create_returns_object: false,
		debug:                 debug,
	}
	client, err := NewAPIClient(opt)
	if err != nil {
		t.Fatal(err)
	}
	client.send_request("POST", "/api/objects", `{ "id": "1234", "first": "Foo", "last": "Bar" }`)

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
