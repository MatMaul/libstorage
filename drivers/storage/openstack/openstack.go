package openstack

import "github.com/akutz/gofig"

const (
	// Name is the provider's name.
	Name = "openstack"
)

func init() {
	registerConfig()
}

func registerConfig() {
	r := gofig.NewRegistration("OpenStack")
	r.Key(gofig.String, "", "", "", "openstack.authURL")
	r.Key(gofig.String, "", "", "", "openstack.userID")
	r.Key(gofig.String, "", "", "", "openstack.userName")
	r.Key(gofig.String, "", "", "", "openstack.password")
	r.Key(gofig.String, "", "", "", "openstack.tenantID")
	r.Key(gofig.String, "", "", "", "openstack.tenantName")
	r.Key(gofig.String, "", "", "", "openstack.domainID")
	r.Key(gofig.String, "", "", "", "openstack.domainName")
	r.Key(gofig.String, "", "", "", "openstack.regionName")
	r.Key(gofig.String, "", "", "", "openstack.availabilityZoneName")
	gofig.Register(r)
}
