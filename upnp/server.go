package upnp

import (
	"mapserver/app"
	"net"
	"time"
	"fmt"
	"log"
	"errors"
	"net/url"
	"net/http"
	"strings"
	"math/rand"
)

const(
	rootDescPath       = "/rootDesc.xml"
	rootDeviceType     = "urn:evidenceb:device:Mapserver:1"
	ssdpInterfaceFlags = net.FlagUp | net.FlagMulticast
)

type UpnpServer struct {
	HTTPConn       net.Listener
	Interfaces     []net.Interface
	UUID           string
	Port           int
	IP             net.IP
}

// https://github.com/anacrolix/dms/blob/master/dlna/dms/dms.go

func generateUUID() string {
	b:= make([]byte, 16)
	rand.Seed(time.Now().UTC().UnixNano())
	rand.Read(b)
	b[8] = b[8] & 0x0f | 0x40 // Version 4 : Random UUID
	b[10] = b[10] & 0x3f | 0x80 // Variant 1 : ?

	return fmt.Sprintf(
		"%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7],
		b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15])
}

func (this *UpnpServer)getXML() []byte {
	url := url.URL{
		Scheme: "http",
		Host: (&net.TCPAddr{
			IP:   this.IP,
			Port: this.Port,
		}).String(),
		Path: "/",
	}
	return []byte("<?xml version=\"1.0\"?>" +
		"<root xmlns=\"urn:schemas-upnp-org:device-1-0\" xmlns:kc=\"urn:schemas-kidscode-org:server-1-0\">" +
		"<specVersion><major>1</major><minor>0</minor></specVersion>" +
		"<device>" +
			"<deviceType>" + rootDeviceType + "</deviceType>" +
			"<friendlyName>Kidscode</friendlyName>" +
			"<manufacturer>EvidenceB</manufacturer>" +
			"<modelName>Kidscode</modelName>" +
			"<UDN>uuid:" + this.UUID + "</UDN>" +
		"</device>" +
		// Kidscode specific part
		"<kc:mapserver>" +
			"<kc:url>" + url.String() + "</kc:url>" +
			"<kc:name>" + "" + "</kc:name>" +
			"<kc:version>" + "" + "</kc:version>" +
		"</kc:server>" +
	"</root>")
}


func (this *UpnpServer)getLocation(ip net.IP) string {
	url := url.URL{
		Scheme: "http",
		Host: (&net.TCPAddr{
			IP:   ip,
			Port: this.HTTPConn.Addr().(*net.TCPAddr).Port,
		}).String(),
		Path: rootDescPath,
	}
	return url.String()
}

func (this *UpnpServer)serveSSDP(intf net.Interface) {
	s := Server{
		Interface: intf,
		Devices:   []string{ rootDeviceType },
		Services:  []string{}, // No service in Kidscode
		Location: func(ip net.IP) string {
			return this.getLocation(ip)
		},
		UUID:           this.UUID,
		NotifyInterval: 30*time.Second,
	}
	if err := s.Init(); err != nil {
		if intf.Flags&ssdpInterfaceFlags != ssdpInterfaceFlags {
			// Didn't expect it to work anyway.
			return
		}
		if strings.Contains(err.Error(), "listen") {
			// OSX has a lot of dud interfaces. Failure to create a socket on
			// the interface are what we're expecting if the interface is no
			// good.
			return
		}
		log.Printf("error creating ssdp server on %s: %s", intf.Name, err)
		return
	}
	defer s.Close()
	log.Println("UPNP: Announce on", intf.Name)
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		if err := s.Serve(); err != nil {
			log.Printf("%q: %q\n", intf.Name, err)
		}
	}()
	select {
//	case <-me.closed:
		// Returning will close the server.
	case <-stopped:
	}
}

func (this *UpnpServer) serveHTTP() error {

	xml := this.getXML()
		//this.HTTPConn.Addr().String())
	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Ext", "")
			w.Header().Set("content-type", `text/xml; charset="utf-8"`)
			w.Header().Set("content-length", fmt.Sprint(len(xml)))
			w.Write(xml)
		}),
	}
	err := srv.Serve(this.HTTPConn)
	select {
//	case <-this.closed:
//		return nil
	default:
		return err
	}
}

func (this *UpnpServer) Run(ctx *app.App) (err error) {
	this.HTTPConn, _ = net.Listen("tcp", "")
	this.UUID = generateUUID()
	this.Port = ctx.Config.Port
	this.IP = nil

	itfs, _ := net.Interfaces()

	// Announce on all non loopback interfaces
	for _, itf := range itfs {
		if (itf.Flags&net.FlagLoopback == net.FlagLoopback) {
			continue;
		}
		// Determine IP adress
		addrs, _ := itf.Addrs()
		var ip net.IP
		for _, addr := range addrs {

			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// Only IPv4
			if (ip != nil && ip.To4() != nil) {
				// Arbitrary use one of the IP as mapserver IP
				// Not satisfying. Maybe it would be better, if possible
				// that HTTP server responds according to the IP of the query
				this.IP = ip.To4()
			}
		}

		go func() {
			this.serveSSDP(itf)
		} ();
	}

	if (this.IP == nil) {
		log.Printf("No IP found, cant start UPNP HTTP server.")
		return errors.New("no IP found")
	} else {
		log.Println("UPNP: Starting HTTP server")
		return this.serveHTTP()
	}
}

func Announce(ctx *app.App) {
	server := new(UpnpServer)
	server.Run(ctx)
}
