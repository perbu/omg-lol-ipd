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
	slog.Info("ExternalIpV4", "externalAddr", externalAddr)
	client, err := getClient(c)
	if err != nil {
		return fmt.Errorf("getClient: %w", err)
	}
	lolAddr, err := getLolAddr(client, c.Username)
	if err != nil {
		return fmt.Errorf("getLolAddr: %w", err)
	}
	if lolAddr.Data != externalAddr {
		slog.Info("lolAddr != externalAddr", "lolAddr", lolAddr, "externalAddr", externalAddr)
		newlol, err := updateLolAddr(client, lolAddr, externalAddr, c.Username)
		if err != nil {
			return fmt.Errorf("setLolAddr: %w", err)
		}
		lolAddr = *newlol
	}
	slog.Info("Starting monitor")
	refreshTicker := time.NewTicker(30 * time.Minute)
	for {
		select {
		case <-refreshTicker.C:
			newAddr, err := ip.ExternalIpV4()
			if err != nil {
				slog.Info("ip.ExternalIpV4 returned an error", "error", err)
				continue
			}
			if newAddr != externalAddr {
				slog.Info("ExternalIpV4", "externalAddr", newAddr)
				newlol, err := updateLolAddr(client, lolAddr, newAddr, c.Username)
				if err != nil {
					slog.Info("setLolAddr returned an error", "error", err)
					continue
				}
				externalAddr = newAddr
				lolAddr = *newlol
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
// it creates an omglol.DNSEntry from the given omglol.DNSRecord and calls
// omglol.Client.UpdateDNSRecord.
func updateLolAddr(c *omglol.Client, record omglol.DNSRecord, newName, domain string) (*omglol.DNSRecord, error) {
	entry := omglol.DNSEntry{
		Type: &record.Type,
		Name: &record.Name,
		Data: &newName,
		TTL:  &record.TTL,
	}
	rec, err := c.UpdateDNSRecord(domain, entry, record.ID)
	if err != nil {
		return nil, fmt.Errorf("c.UpdateDNSRecord: %w", err)
	}
	return rec, nil

}

func getLolAddr(c *omglol.Client, hostname string) (omglol.DNSRecord, error) {
	null := omglol.DNSRecord{}
	addrs, err := c.ListDNSRecords(hostname)
	if err != nil {
		return null, fmt.Errorf("c.ListDNSRecords: %w", err)
	}
	if len(*addrs) == 0 {
		return null, fmt.Errorf("no addresses found")
	}
	for _, addr := range *addrs {
		if addr.Type == "A" {
			return addr, nil
		}
	}
	return null, fmt.Errorf("no A record found")
}
