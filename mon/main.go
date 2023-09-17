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

// Monitor continuously checks and updates the DNS record if the external IP changes.
func Monitor(ctx context.Context, c config.Config) error {
	client, err := getClient(c)
	if err != nil {
		return err
	}
	externalAddr, err := ip.ExternalIpV4()
	if err != nil {
		return fmt.Errorf("retrieving external IP: %w", err)
	}

	id, existing, err := getLolIdName(client, c.Username, c.Hostname)
	if err != nil {
		return err
	}
	if existing != externalAddr {
		err = updateDNS(client, id, c, externalAddr)
		if err != nil {
			return err
		}
	}
	return monitorIPChanges(ctx, client, id, c, externalAddr)
}

func getClient(c config.Config) (*omglol.Client, error) {
	return omglol.NewClient(c.Email, c.ApiKey)
}

func updateDNS(c *omglol.Client, id int64, configData config.Config, externalAddr string) error {
	err := updateLolAddr(c, id, configData.Hostname, externalAddr, configData.Username)
	if err != nil {
		return fmt.Errorf("updating DNS record: %w", err)
	}
	slog.Info("Updated DNS record", "hostname", configData.Hostname, "username", configData.Username, "externalAddr", externalAddr)
	return nil
}

func monitorIPChanges(ctx context.Context, client *omglol.Client, id int64, c config.Config, currentAddr string) error {
	refreshTicker := time.NewTicker(30 * time.Minute)
	for {
		select {
		case <-refreshTicker.C:
			newAddr, err := ip.ExternalIpV4()
			if err != nil {
				slog.Warn("retrieving external IP", "error", err)
				continue
			}
			if newAddr != currentAddr {
				err := updateDNS(client, id, c, newAddr)
				if err != nil {
					slog.Warn("failed to update DNS record", "error", err)
					continue
				}
				currentAddr = newAddr
			}
		case <-ctx.Done():
			return nil
		}
	}
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
