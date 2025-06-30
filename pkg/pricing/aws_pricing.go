package pricing

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// PricingData represents AWS pricing information for an instance type
type PricingData struct {
	InstanceType string  `json:"instance_type"`
	Region       string  `json:"region"`
	OnDemand     float64 `json:"on_demand_hourly"`
	Currency     string  `json:"currency"`
	LastUpdated  string  `json:"last_updated"`
}

// PricingService handles AWS pricing information
type PricingService struct {
	client *http.Client
}

// NewPricingService creates a new pricing service
func NewPricingService() *PricingService {
	return &PricingService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetInstancePricing retrieves current pricing for an instance type in a specific region
func (p *PricingService) GetInstancePricing(ctx context.Context, instanceType, region string) (*PricingData, error) {
	// AWS Pricing API endpoint (simplified for demo - real implementation would use AWS SDK)
	// For now, we'll use a hardcoded pricing table based on current AWS pricing
	price, err := p.getHardcodedPricing(instanceType, region)
	if err != nil {
		return nil, err
	}

	return &PricingData{
		InstanceType: instanceType,
		Region:       region,
		OnDemand:     price,
		Currency:     "USD",
		LastUpdated:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// getHardcodedPricing returns hardcoded pricing based on AWS pricing as of June 2025
func (p *PricingService) getHardcodedPricing(instanceType, region string) (float64, error) {
	// Pricing table for us-east-1 (other regions have multipliers)
	basePricing := map[string]float64{
		// M7i instances (Intel Ice Lake)
		"m7i.large":    0.1008,
		"m7i.xlarge":   0.2016,
		"m7i.2xlarge":  0.4032,
		"m7i.4xlarge":  0.8064,
		"m7i.8xlarge":  1.6128,
		"m7i.12xlarge": 2.4192,
		"m7i.16xlarge": 3.2256,
		"m7i.24xlarge": 4.8384,
		"m7i.48xlarge": 9.6768,

		// M7a instances (AMD EPYC)
		"m7a.large":    0.0864,
		"m7a.xlarge":   0.1728,
		"m7a.2xlarge":  0.3456,
		"m7a.4xlarge":  0.6912,
		"m7a.8xlarge":  1.3824,
		"m7a.12xlarge": 2.0736,
		"m7a.16xlarge": 2.7648,
		"m7a.24xlarge": 4.1472,
		"m7a.48xlarge": 8.2944,

		// M7g instances (Graviton3)
		"m7g.large":    0.0808,
		"m7g.xlarge":   0.1616,
		"m7g.2xlarge":  0.3232,
		"m7g.4xlarge":  0.6464,
		"m7g.8xlarge":  1.2928,
		"m7g.12xlarge": 1.9392,
		"m7g.16xlarge": 2.5856,

		// C7i instances (Intel Ice Lake - Compute optimized)
		"c7i.large":    0.0850,
		"c7i.xlarge":   0.1700,
		"c7i.2xlarge":  0.3400,
		"c7i.4xlarge":  0.6800,
		"c7i.8xlarge":  1.3600,
		"c7i.12xlarge": 2.0400,
		"c7i.16xlarge": 2.7200,
		"c7i.24xlarge": 4.0800,
		"c7i.48xlarge": 8.1600,

		// C7a instances (AMD EPYC - Compute optimized)
		"c7a.large":    0.0765,
		"c7a.xlarge":   0.1530,
		"c7a.2xlarge":  0.3060,
		"c7a.4xlarge":  0.6120,
		"c7a.8xlarge":  1.2240,
		"c7a.12xlarge": 1.8360,
		"c7a.16xlarge": 2.4480,
		"c7a.24xlarge": 3.6720,
		"c7a.48xlarge": 7.3440,

		// C7g instances (Graviton3 - Compute optimized)
		"c7g.medium":   0.0362,
		"c7g.large":    0.0725,
		"c7g.xlarge":   0.1450,
		"c7g.2xlarge":  0.2900,
		"c7g.4xlarge":  0.5800,
		"c7g.8xlarge":  1.1600,
		"c7g.12xlarge": 1.7400,
		"c7g.16xlarge": 2.3200,

		// R7i instances (Intel Ice Lake - Memory optimized)
		"r7i.large":    0.1512,
		"r7i.xlarge":   0.3024,
		"r7i.2xlarge":  0.6048,
		"r7i.4xlarge":  1.2096,
		"r7i.8xlarge":  2.4192,
		"r7i.12xlarge": 3.6288,
		"r7i.16xlarge": 4.8384,
		"r7i.24xlarge": 7.2576,
		"r7i.48xlarge": 14.5152,

		// R7a instances (AMD EPYC - Memory optimized)
		"r7a.large":    0.1260,
		"r7a.xlarge":   0.2520,
		"r7a.2xlarge":  0.5040,
		"r7a.4xlarge":  1.0080,
		"r7a.8xlarge":  2.0160,
		"r7a.12xlarge": 3.0240,
		"r7a.16xlarge": 4.0320,
		"r7a.24xlarge": 6.0480,
		"r7a.48xlarge": 12.0960,

		// R7g instances (Graviton3 - Memory optimized)
		"r7g.large":    0.1344,
		"r7g.xlarge":   0.2688,
		"r7g.2xlarge":  0.5376,
		"r7g.4xlarge":  1.0752,
		"r7g.8xlarge":  2.1504,
		"r7g.12xlarge": 3.2256,
		"r7g.16xlarge": 4.3008,
	}

	basePrice, exists := basePricing[instanceType]
	if !exists {
		return 0, fmt.Errorf("pricing not available for instance type %s", instanceType)
	}

	// Apply regional multipliers
	regionalMultiplier := getRegionalMultiplier(region)
	return basePrice * regionalMultiplier, nil
}

// getRegionalMultiplier returns pricing multiplier for different regions
func getRegionalMultiplier(region string) float64 {
	regionalMultipliers := map[string]float64{
		"us-east-1":      1.0,    // Base pricing
		"us-east-2":      1.0,    // Same as us-east-1
		"us-west-1":      1.05,   // 5% higher
		"us-west-2":      1.0,    // Same as us-east-1
		"eu-west-1":      1.08,   // 8% higher
		"eu-west-2":      1.10,   // 10% higher
		"eu-central-1":   1.12,   // 12% higher
		"ap-southeast-1": 1.15,   // 15% higher
		"ap-southeast-2": 1.18,   // 18% higher
		"ap-northeast-1": 1.20,   // 20% higher
	}

	if multiplier, exists := regionalMultipliers[region]; exists {
		return multiplier
	}
	return 1.1 // Default 10% higher for unknown regions
}

// PerformanceMetrics represents performance data for price/performance calculations
type PerformanceMetrics struct {
	TriadBandwidth float64 `json:"triad_bandwidth"`
	CopyBandwidth  float64 `json:"copy_bandwidth"`
	ScaleBandwidth float64 `json:"scale_bandwidth"`
	AddBandwidth   float64 `json:"add_bandwidth"`
	GFLOPS         float64 `json:"gflops,omitempty"`
	Efficiency     float64 `json:"efficiency,omitempty"`
}

// PricePerformanceMetrics represents calculated price/performance ratios
type PricePerformanceMetrics struct {
	InstanceType           string  `json:"instance_type"`
	Region                 string  `json:"region"`
	HourlyPrice           float64 `json:"hourly_price"`
	TriadBandwidth        float64 `json:"triad_bandwidth"`
	PricePerGBps          float64 `json:"price_per_gbps"`           // $/hour per GB/s
	NormalizedScore       float64 `json:"normalized_score"`        // Normalized to baseline
	BaselineInstance      string  `json:"baseline_instance"`       // Reference instance
	BaselineScore         float64 `json:"baseline_score"`          // Reference score
	PerformanceRatio      float64 `json:"performance_ratio"`       // vs baseline performance
	CostEfficiencyRatio   float64 `json:"cost_efficiency_ratio"`   // vs baseline cost efficiency
	ValueScore            float64 `json:"value_score"`             // Combined performance/cost score
}

// PricePerformanceCalculator calculates price/performance metrics
type PricePerformanceCalculator struct {
	pricingService *PricingService
	baseline       *PricePerformanceMetrics // Reference point for normalization
}

// NewPricePerformanceCalculator creates a new calculator with baseline
func NewPricePerformanceCalculator(baseline *PricePerformanceMetrics) *PricePerformanceCalculator {
	return &PricePerformanceCalculator{
		pricingService: NewPricingService(),
		baseline:       baseline,
	}
}

// CalculatePricePerformance calculates comprehensive price/performance metrics
func (calc *PricePerformanceCalculator) CalculatePricePerformance(
	ctx context.Context,
	instanceType, region string,
	metrics *PerformanceMetrics,
) (*PricePerformanceMetrics, error) {
	
	// Get pricing data
	pricing, err := calc.pricingService.GetInstancePricing(ctx, instanceType, region)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing: %w", err)
	}

	// Calculate basic price/performance ratio
	pricePerGBps := pricing.OnDemand / metrics.TriadBandwidth

	// Calculate normalized scores against baseline
	var normalizedScore, performanceRatio, costEfficiencyRatio, valueScore float64
	var baselineInstance string
	var baselineScore float64

	if calc.baseline != nil {
		baselineInstance = calc.baseline.InstanceType
		baselineScore = calc.baseline.PricePerGBps
		
		// Performance ratio: how much better/worse performance is vs baseline
		performanceRatio = metrics.TriadBandwidth / calc.baseline.TriadBandwidth
		
		// Cost efficiency ratio: how much better/worse cost efficiency is vs baseline
		costEfficiencyRatio = calc.baseline.PricePerGBps / pricePerGBps
		
		// Normalized score: 1.0 = same as baseline, >1.0 = better value
		normalizedScore = costEfficiencyRatio
		
		// Value score: combined performance and cost efficiency
		// Higher performance = good, lower cost per unit = good
		valueScore = performanceRatio * costEfficiencyRatio
	}

	return &PricePerformanceMetrics{
		InstanceType:        instanceType,
		Region:             region,
		HourlyPrice:        pricing.OnDemand,
		TriadBandwidth:     metrics.TriadBandwidth,
		PricePerGBps:       pricePerGBps,
		NormalizedScore:    normalizedScore,
		BaselineInstance:   baselineInstance,
		BaselineScore:      baselineScore,
		PerformanceRatio:   performanceRatio,
		CostEfficiencyRatio: costEfficiencyRatio,
		ValueScore:         valueScore,
	}, nil
}

// GetDefaultBaseline returns a reasonable baseline instance for normalization
func GetDefaultBaseline(ctx context.Context) (*PricePerformanceMetrics, error) {
	// Use m7i.large as baseline - common instance with good balance
	calc := &PricePerformanceCalculator{
		pricingService: NewPricingService(),
	}
	
	pricing, err := calc.pricingService.GetInstancePricing(ctx, "m7i.large", "us-east-1")
	if err != nil {
		return nil, err
	}

	// Typical performance for m7i.large (from our data)
	baselineMetrics := &PerformanceMetrics{
		TriadBandwidth: 41.9, // GB/s
	}

	return &PricePerformanceMetrics{
		InstanceType:   "m7i.large",
		Region:        "us-east-1",
		HourlyPrice:   pricing.OnDemand,
		TriadBandwidth: baselineMetrics.TriadBandwidth,
		PricePerGBps:  pricing.OnDemand / baselineMetrics.TriadBandwidth,
	}, nil
}

// BatchCalculatePricePerformance calculates metrics for multiple instances
func (calc *PricePerformanceCalculator) BatchCalculatePricePerformance(
	ctx context.Context,
	instances []struct {
		InstanceType string
		Region       string
		Metrics      *PerformanceMetrics
	},
) ([]*PricePerformanceMetrics, error) {
	
	results := make([]*PricePerformanceMetrics, 0, len(instances))
	
	for _, instance := range instances {
		result, err := calc.CalculatePricePerformance(
			ctx,
			instance.InstanceType,
			instance.Region,
			instance.Metrics,
		)
		if err != nil {
			// Log error but continue with other instances
			fmt.Printf("Warning: failed to calculate price/performance for %s: %v\n", 
				instance.InstanceType, err)
			continue
		}
		
		results = append(results, result)
	}
	
	return results, nil
}