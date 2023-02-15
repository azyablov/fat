package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/azyablov/fat/srl/v22.6.4/system"
	"github.com/openconfig/gnmic/api"
	"github.com/openconfig/ygot/ygot"
	"google.golang.org/protobuf/encoding/prototext"
)

func main() {
	srl := new(system.Device)
	srl.System = new(system.SrlNokiaSystem_System)
	srl.System.Name = new(system.SrlNokiaSystem_System_Name)
	srl.System.Dns = new(system.SrlNokiaSystem_System_Dns)
	srl.System.Name.HostName = ygot.String("ygot-pg")
	srl.System.Name.DomainName = ygot.String("ygot.com")

	srl.System.Dns.ServerList = append(srl.System.Dns.ServerList, "172.22.1.1")

	srl.System.Dns.AppendHostEntry(&system.SrlNokiaSystem_System_Dns_HostEntry{
		Name:        ygot.String("hv2"),
		Ipv4Address: ygot.String("172.22.1.2")})

	hv2, err := srl.System.Dns.NewHostEntry("hv3")
	if err != nil {
		log.Panicf("fail to create new host entry:  %s", err)
	}
	hv2.Ipv4Address = ygot.String("172.22.1.3")
	if err := srl.System.Validate(); err != nil {
		log.Panicf("unable to validate: %s", err)
	}

	mapJSON, err := ygot.ConstructIETFJSON(srl, nil)
	if err != nil {
		log.Panicf("unable to marshall into map:  %s", err)
	}
	js, err := json.MarshalIndent(mapJSON, "", "    ")
	fmt.Println(string(js))

	// gNMI

	t, err := api.NewTarget(api.Name("clab-2nd-srl1"),
		api.Address("clab-2nd-srl1:57400"),
		api.Username("admin"),
		api.Password("admin"),
		api.Timeout(time.Second*5),
		api.SkipVerify(true),
		api.TLSMinVersion("1.2"))
	if err != nil {
		log.Panicf("can't create new target: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = t.CreateGNMIClient(ctx)
	defer t.Close()
	if err != nil {
		log.Panicf("unable create gNMI client: %s", err)
	}

	getSystem, err := api.NewGetRequest(api.Path("/system"),
		api.Prefix("/"),
		api.DataTypeCONFIG(),
		api.EncodingJSON_IETF())

	if err != nil {
		log.Panicf("can't construct gNMI Get request: %s", err)
	}

	fmt.Println(prototext.Format(getSystem))

	respSystem, err := t.Get(ctx, getSystem)
	if err != nil {
		log.Panicf("error sending gNMI Get request: %s", err)
	}
	fmt.Println(prototext.Format(respSystem))
}
