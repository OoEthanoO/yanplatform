// Data models mirroring the Go backend structures.

class Resource {
  final String id;
  final String name;
  final List<String> hsCodes;
  final String primaryRegion;

  Resource({
    required this.id,
    required this.name,
    required this.hsCodes,
    required this.primaryRegion,
  });

  factory Resource.fromJson(Map<String, dynamic> json) {
    return Resource(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      hsCodes: List<String>.from(json['hs_codes'] ?? []),
      primaryRegion: json['primary_region'] ?? '',
    );
  }
}

class RiskScore {
  final String id;
  final String region;
  final String country;
  final String resource;
  final double overallScore;
  final double supplyConcentration;
  final double geopoliticalTension;
  final double tradePolicySignal;
  final double logisticsRisk;
  final String computedAt;
  final bool isHighRisk;

  RiskScore({
    required this.id,
    required this.region,
    required this.country,
    required this.resource,
    required this.overallScore,
    required this.supplyConcentration,
    required this.geopoliticalTension,
    required this.tradePolicySignal,
    required this.logisticsRisk,
    required this.computedAt,
    required this.isHighRisk,
  });

  factory RiskScore.fromJson(Map<String, dynamic> json) {
    return RiskScore(
      id: json['id'] ?? '',
      region: json['region'] ?? '',
      country: json['country'] ?? '',
      resource: json['resource'] ?? '',
      overallScore: (json['overall_score'] ?? 0).toDouble(),
      supplyConcentration: (json['supply_concentration'] ?? 0).toDouble(),
      geopoliticalTension: (json['geopolitical_tension'] ?? 0).toDouble(),
      tradePolicySignal: (json['trade_policy_signal'] ?? 0).toDouble(),
      logisticsRisk: (json['logistics_risk'] ?? 0).toDouble(),
      computedAt: json['computed_at'] ?? '',
      isHighRisk: json['is_high_risk'] ?? false,
    );
  }
}

class RiskOverview {
  final Map<String, RiskScore> resourceRisks;
  final int recentEvents;
  final int highRiskZones;
  final String lastUpdated;

  RiskOverview({
    required this.resourceRisks,
    required this.recentEvents,
    required this.highRiskZones,
    required this.lastUpdated,
  });

  factory RiskOverview.fromJson(Map<String, dynamic> json) {
    final risksMap = <String, RiskScore>{};
    if (json['resource_risks'] != null) {
      (json['resource_risks'] as Map<String, dynamic>).forEach((key, value) {
        risksMap[key] = RiskScore.fromJson(value);
      });
    }

    return RiskOverview(
      resourceRisks: risksMap,
      recentEvents: json['recent_events'] ?? 0,
      highRiskZones: json['high_risk_zones'] ?? 0,
      lastUpdated: json['last_updated'] ?? '',
    );
  }
}

class GDELTEvent {
  final String id;
  final String eventDate;
  final String actor1Name;
  final String actor1Country;
  final String actor2Name;
  final String actor2Country;
  final String eventType;
  final String description;
  final double avgTone;
  final double goldsteinScale;
  final String sourceUrl;
  final double relevance;
  final String sentimentLabel;

  GDELTEvent({
    required this.id,
    required this.eventDate,
    required this.actor1Name,
    required this.actor1Country,
    required this.actor2Name,
    required this.actor2Country,
    required this.eventType,
    required this.description,
    required this.avgTone,
    required this.goldsteinScale,
    required this.sourceUrl,
    required this.relevance,
    required this.sentimentLabel,
  });

  factory GDELTEvent.fromJson(Map<String, dynamic> json) {
    return GDELTEvent(
      id: json['id'] ?? '',
      eventDate: json['event_date'] ?? '',
      actor1Name: json['actor1_name'] ?? '',
      actor1Country: json['actor1_country'] ?? '',
      actor2Name: json['actor2_name'] ?? '',
      actor2Country: json['actor2_country'] ?? '',
      eventType: json['event_type'] ?? '',
      description: json['description'] ?? '',
      avgTone: (json['avg_tone'] ?? 0).toDouble(),
      goldsteinScale: (json['goldstein_scale'] ?? 0).toDouble(),
      sourceUrl: json['source_url'] ?? '',
      relevance: (json['relevance'] ?? 0).toDouble(),
      sentimentLabel: json['sentiment_label'] ?? 'neutral',
    );
  }
}

class Chokepoint {
  final String id;
  final String name;
  final String type;
  final String country;
  final String region;
  final double globalSharePct;
  final String resource;
  final String riskLevel;
  final double latitude;
  final double longitude;

  Chokepoint({
    required this.id,
    required this.name,
    required this.type,
    required this.country,
    required this.region,
    required this.globalSharePct,
    required this.resource,
    required this.riskLevel,
    required this.latitude,
    required this.longitude,
  });

  factory Chokepoint.fromJson(Map<String, dynamic> json) {
    return Chokepoint(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      type: json['type'] ?? '',
      country: json['country'] ?? '',
      region: json['region'] ?? '',
      globalSharePct: (json['global_share_pct'] ?? 0).toDouble(),
      resource: json['resource'] ?? '',
      riskLevel: json['risk_level'] ?? 'low',
      latitude: (json['latitude'] ?? 0).toDouble(),
      longitude: (json['longitude'] ?? 0).toDouble(),
    );
  }
}

class RerouteResult {
  final String id;
  final String triggerRegion;
  final double triggerRiskScore;
  final String resource;
  final List<RerouteAlternative> alternatives;
  final String simulatedAt;

  RerouteResult({
    required this.id,
    required this.triggerRegion,
    required this.triggerRiskScore,
    required this.resource,
    required this.alternatives,
    required this.simulatedAt,
  });

  factory RerouteResult.fromJson(Map<String, dynamic> json) {
    return RerouteResult(
      id: json['id'] ?? '',
      triggerRegion: json['trigger_region'] ?? '',
      triggerRiskScore: (json['trigger_risk_score'] ?? 0).toDouble(),
      resource: json['resource'] ?? '',
      alternatives: (json['alternatives'] as List<dynamic>?)
              ?.map((a) => RerouteAlternative.fromJson(a))
              .toList() ??
          [],
      simulatedAt: json['simulated_at'] ?? '',
    );
  }
}

class RerouteAlternative {
  final String supplierId;
  final String supplierName;
  final String country;
  final double capacityTonnes;
  final double absorptionPct;
  final double feasibilityScore;
  final int leadTimeDays;
  final double latitude;
  final double longitude;

  RerouteAlternative({
    required this.supplierId,
    required this.supplierName,
    required this.country,
    required this.capacityTonnes,
    required this.absorptionPct,
    required this.feasibilityScore,
    required this.leadTimeDays,
    required this.latitude,
    required this.longitude,
  });

  factory RerouteAlternative.fromJson(Map<String, dynamic> json) {
    return RerouteAlternative(
      supplierId: json['supplier_id'] ?? '',
      supplierName: json['supplier_name'] ?? '',
      country: json['country'] ?? '',
      capacityTonnes: (json['capacity_tonnes'] ?? 0).toDouble(),
      absorptionPct: (json['absorption_pct'] ?? 0).toDouble(),
      feasibilityScore: (json['feasibility_score'] ?? 0).toDouble(),
      leadTimeDays: json['lead_time_days'] ?? 0,
      latitude: (json['latitude'] ?? 0).toDouble(),
      longitude: (json['longitude'] ?? 0).toDouble(),
    );
  }
}

class Supplier {
  final String id;
  final String name;
  final String country;
  final String region;
  final String resource;
  final double capacityTonnesYr;
  final double neutralityScore;
  final double latitude;
  final double longitude;
  final bool isAlternative;

  Supplier({
    required this.id,
    required this.name,
    required this.country,
    required this.region,
    required this.resource,
    required this.capacityTonnesYr,
    required this.neutralityScore,
    required this.latitude,
    required this.longitude,
    required this.isAlternative,
  });

  factory Supplier.fromJson(Map<String, dynamic> json) {
    return Supplier(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      country: json['country'] ?? '',
      region: json['region'] ?? '',
      resource: json['resource'] ?? '',
      capacityTonnesYr: (json['capacity_tonnes_yr'] ?? 0).toDouble(),
      neutralityScore: (json['neutrality_score'] ?? 0).toDouble(),
      latitude: (json['latitude'] ?? 0).toDouble(),
      longitude: (json['longitude'] ?? 0).toDouble(),
      isAlternative: json['is_alternative'] ?? false,
    );
  }
}

class TradeFlow {
  final String id;
  final int year;
  final int month;
  final String reporterCountry;
  final String partnerCountry;
  final String hsCode;
  final String resource;
  final String flowType;
  final double valueUsd;
  final double weightKg;

  TradeFlow({
    required this.id,
    required this.year,
    required this.month,
    required this.reporterCountry,
    required this.partnerCountry,
    required this.hsCode,
    required this.resource,
    required this.flowType,
    required this.valueUsd,
    required this.weightKg,
  });

  factory TradeFlow.fromJson(Map<String, dynamic> json) {
    return TradeFlow(
      id: json['id'] ?? '',
      year: json['year'] ?? 0,
      month: json['month'] ?? 0,
      reporterCountry: json['reporter_country'] ?? '',
      partnerCountry: json['partner_country'] ?? '',
      hsCode: json['hs_code'] ?? '',
      resource: json['resource'] ?? '',
      flowType: json['flow_type'] ?? '',
      valueUsd: (json['value_usd'] ?? 0).toDouble(),
      weightKg: (json['weight_kg'] ?? 0).toDouble(),
    );
  }
}
