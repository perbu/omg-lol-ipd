package mon

import (
	"context"
	"fmt"
	"github.com/ejstreet/omglol-client-go/omglol"
	"github.com/perbu/omg-lol-ipd/config"
	"github.com/perbu/omg-lol-ipd/ip"
	"log/slog"
	"time"
)

func Monitor(ctx context.Context, c config.Config) error {
	externalAddr, err := ip.ExternalIpV4()
	if err != nil {
		return fmt.Errorf("ip.ExternalIpV4: %w", err)
	}
	client, err := getClient(c)
	if err != nil {
		return fmt.Errorf("getClient: %w", err)
	}
	id, existing, err := getLolIdName(client, c.Username, c.Hostname)
	if err != nil {
		return fmt.Errorf("getLolAddr: %w", err)
	}
	if existing != externalAddr {
		err := updateLolAddr(client, id, c.Hostname, externalAddr, c.Username)
		if err != nil {
			return fmt.Errorf("updateLolAddr: %w", err)
		}
		slog.Info("Updated DNS record", "hostname", c.Hostname, "username", c.Username, "externalAddr", externalAddr)
	}
	refreshTicker := time.NewTicker(30 * time.Minute)
	for {

		select {
		case <-refreshTicker.C:
			newAddr, err := ip.ExternalIpV4()
			if err != nil {
				slog.Warn("ip.ExternalIpV4 returned an error", "error", err)
				continue
			}
			if newAddr != externalAddr {
				slog.Info("Updating DNS record", "hostname", c.Hostname, "username", c.Username, "externalAddr", newAddr)
				err := updateLolAddr(client, id, c.Hostname, newAddr, c.Username)
				if err != nil {
					slog.Warn("updateLolAddr failed", "error", err)
					continue
				}
				externalAddr = newAddr
			}
		case <-ctx.Done():
			return nil
		}

	}
}

func getClient(c config.Config) (*omglol.Client, error) {
	return omglol.NewClient(c.Email, c.ApiKey)
}

// updateLolAddr updates the DNS record for the given domain with the new name.
func updateLolAddr(c *omglol.Client, id int64, hostname, addr, domain string) error {
	recType := "A"
	ttl := int64(300)
	entry := omglol.DNSEntry{
		Type: &recType,
		Name: &hostname,
		Data: &addr,
		TTL:  &ttl,
	}
	_, err := c.UpdateDNSRecord(domain, entry, id)
	if err != nil {
		return fmt.Errorf("c.UpdateDNSRecord: %w", err)
	}
	return nil

}

func getLolIdName(c *omglol.Client, username, hostname string) (int64, string, error) {
	desired := fmt.Sprintf("%s.%s", hostname, username)
	addrs, err := c.ListDNSRecords(username)
	if err != nil {
		return 0, "", fmt.Errorf("c.ListDNSRecords: %w", err)
	}
	if len(*addrs) == 0 {
		return 0, "", fmt.Errorf("no addresses found")
	}
	for _, addr := range *addrs {
		if addr.Type == "A" && addr.Name == desired {
			return addr.ID, addr.Data, nil

		}
	}
	return 0, "", fmt.Errorf("no A record found for'%s'", desired)
}
