package cron

import (
	"fmt"
	"log"
	"time"

	"github.com/lineserve/lineserve-api/pkg/client"
)

// VPSBillingJob represents a job that runs VPS billing
type VPSBillingJob struct {
	SupabaseClient *client.SupabaseClient
}

// NewVPSBillingJob creates a new VPS billing job
func NewVPSBillingJob(supabaseClient *client.SupabaseClient) *VPSBillingJob {
	return &VPSBillingJob{
		SupabaseClient: supabaseClient,
	}
}

// RunVPSRenewalBilling runs the VPS renewal billing process
func (j *VPSBillingJob) RunVPSRenewalBilling() error {
	log.Println("Running VPS renewal billing...")

	// Run renewal billing
	results, err := j.SupabaseClient.RunVPSRenewalBilling()
	if err != nil {
		return fmt.Errorf("failed to run renewal billing: %v", err)
	}

	// Log results
	log.Printf("VPS renewal billing completed. Processed %d subscriptions.", len(results))
	for _, result := range results {
		log.Printf("Subscription %s (User: %s, Plan: %s): %s",
			result.SubscriptionID, result.UserID, result.PlanCode, result.RenewalResult)
	}

	return nil
}

// StartVPSBillingCron starts the VPS billing cron job
func StartVPSBillingCron(supabaseClient *client.SupabaseClient) {
	job := NewVPSBillingJob(supabaseClient)

	// Run immediately on startup
	if err := job.RunVPSRenewalBilling(); err != nil {
		log.Printf("Error running VPS renewal billing: %v", err)
	}

	// Run daily at midnight
	go func() {
		for {
			// Calculate time until next run (midnight)
			now := time.Now()
			nextRun := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			duration := nextRun.Sub(now)

			log.Printf("Next VPS billing run scheduled for %s (in %s)", nextRun.Format(time.RFC3339), duration)

			// Sleep until next run
			time.Sleep(duration)

			// Run billing
			if err := job.RunVPSRenewalBilling(); err != nil {
				log.Printf("Error running VPS renewal billing: %v", err)
			}
		}
	}()
}
