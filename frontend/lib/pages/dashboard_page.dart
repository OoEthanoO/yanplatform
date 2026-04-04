import 'package:flutter/material.dart';
import '../models/models.dart';
import '../services/api_service.dart';

/// Main dashboard page with risk overview, events, chokepoints, and trade data.
class DashboardPage extends StatefulWidget {
  final ApiService apiService;

  const DashboardPage({super.key, required this.apiService});

  @override
  State<DashboardPage> createState() => _DashboardPageState();
}

class _DashboardPageState extends State<DashboardPage> {
  RiskOverview? _overview;
  List<GDELTEvent> _events = [];
  List<Chokepoint> _chokepoints = [];
  List<TradeFlow> _tradeFlows = [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    setState(() {
      _loading = true;
      _error = null;
    });

    try {
      final results = await Future.wait([
        widget.apiService.getRiskOverview(),
        widget.apiService.getRecentEvents(),
        widget.apiService.getChokepoints(),
        widget.apiService.getTradeFlows(),
      ]);

      setState(() {
        _overview = results[0] as RiskOverview;
        _events = results[1] as List<GDELTEvent>;
        _chokepoints = results[2] as List<Chokepoint>;
        _tradeFlows = results[3] as List<TradeFlow>;
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Scaffold(
      backgroundColor: colorScheme.surface,
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? _buildErrorState(colorScheme)
              : RefreshIndicator(
                  onRefresh: _loadData,
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.all(24),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        _buildHeader(colorScheme),
                        const SizedBox(height: 24),
                        _buildRiskCards(colorScheme),
                        const SizedBox(height: 24),
                        _buildStatsRow(colorScheme),
                        const SizedBox(height: 24),
                        Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Expanded(
                              flex: 3,
                              child: _buildEventsFeed(colorScheme),
                            ),
                            const SizedBox(width: 24),
                            Expanded(
                              flex: 2,
                              child: _buildChokepointsList(colorScheme),
                            ),
                          ],
                        ),
                        const SizedBox(height: 24),
                        _buildTradeFlowsTable(colorScheme),
                      ],
                    ),
                  ),
                ),
    );
  }

  Widget _buildErrorState(ColorScheme colorScheme) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.cloud_off, size: 64, color: colorScheme.error),
          const SizedBox(height: 16),
          Text(
            'Unable to connect to backend',
            style: TextStyle(
              fontSize: 20,
              fontWeight: FontWeight.w600,
              color: colorScheme.onSurface,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Make sure the Go server is running on localhost:8080',
            style: TextStyle(color: colorScheme.onSurfaceVariant),
          ),
          const SizedBox(height: 8),
          Text(
            _error ?? '',
            style: TextStyle(
              color: colorScheme.error,
              fontSize: 12,
            ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 24),
          FilledButton.icon(
            onPressed: _loadData,
            icon: const Icon(Icons.refresh),
            label: const Text('Retry'),
          ),
        ],
      ),
    );
  }

  Widget _buildHeader(ColorScheme colorScheme) {
    return Row(
      children: [
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'YanPlatform',
                style: TextStyle(
                  fontSize: 28,
                  fontWeight: FontWeight.bold,
                  color: colorScheme.onSurface,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                _overview == null
                    ? 'Scan ongoing...'
                    : '${_overview!.resourceRisks.keys.map((k) => k[0].toUpperCase() + k.substring(1)).join(', ')} — Real-time Risk Monitoring',
                style: TextStyle(
                  fontSize: 14,
                  color: colorScheme.onSurfaceVariant,
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ],
          ),
        ),
        const SizedBox(width: 16),
        FilledButton.tonalIcon(
          onPressed: _loadData,
          icon: const Icon(Icons.refresh, size: 18),
          label: const Text('Refresh Data'),
        ),
      ],
    );
  }

  Widget _buildRiskCards(ColorScheme colorScheme) {
    if (_overview == null) return const SizedBox.shrink();

    final risks = _overview!.resourceRisks.entries.toList();
    if (risks.isEmpty) return const SizedBox.shrink();

    // Group risks into rows of 2
    final rows = <Widget>[];
    for (var i = 0; i < risks.length; i += 2) {
      final isLastSingle = i + 1 == risks.length;
      rows.add(
        Row(
          children: [
            Expanded(
              child: _buildRiskCard(
                _formatResourceName(risks[i].key),
                risks[i].value,
                _getResourceIcon(risks[i].key),
                colorScheme,
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: isLastSingle
                  ? const SizedBox.shrink()
                  : _buildRiskCard(
                      _formatResourceName(risks[i + 1].key),
                      risks[i + 1].value,
                      _getResourceIcon(risks[i + 1].key),
                      colorScheme,
                    ),
            ),
          ],
        ),
      );
      if (i + 2 < risks.length || (i + 1 < risks.length && !isLastSingle)) {
        rows.add(const SizedBox(height: 16));
      }
    }

    return Column(children: rows);
  }

  String _formatResourceName(String key) {
    if (key.isEmpty) return '';
    // e.g. "gallium" -> "Gallium (Ga)"
    final caps = key[0].toUpperCase() + key.substring(1);
    switch (key.toLowerCase()) {
      case 'gallium':
        return 'Gallium (Ga)';
      case 'germanium':
        return 'Germanium (Ge)';
      case 'lithium':
        return 'Lithium (Li)';
      case 'cobalt':
        return 'Cobalt (Co)';
      case 'graphite':
        return 'Graphite (C)';
      default:
        return caps;
    }
  }

  IconData _getResourceIcon(String key) {
    switch (key.toLowerCase()) {
      case 'gallium':
        return Icons.science_outlined;
      case 'germanium':
        return Icons.memory_outlined;
      case 'lithium':
        return Icons.battery_charging_full_outlined;
      case 'cobalt':
        return Icons.bolt_outlined;
      case 'graphite':
        return Icons.layers_outlined;
      default:
        return Icons.public_outlined;
    }
  }

  Widget _buildRiskCard(
      String title, RiskScore risk, IconData icon, ColorScheme colorScheme) {
    final riskColor = _getRiskColor(risk.overallScore);

    return Card(
      color: colorScheme.surfaceContainerHighest,
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(icon, color: colorScheme.primary, size: 24),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    title,
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.w600,
                      color: colorScheme.onSurface,
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                const SizedBox(width: 8),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                  decoration: BoxDecoration(
                    color: riskColor.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(color: riskColor.withValues(alpha: 0.3)),
                  ),
                  child: Text(
                    risk.isHighRisk ? 'HIGH RISK' : 'MODERATE',
                    style: TextStyle(
                      color: riskColor,
                      fontWeight: FontWeight.bold,
                      fontSize: 12,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 20),
            // Overall Score
            Row(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Text(
                  risk.overallScore.toStringAsFixed(0),
                  style: TextStyle(
                    fontSize: 48,
                    fontWeight: FontWeight.bold,
                    color: riskColor,
                    height: 1,
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.only(bottom: 8, left: 4),
                  child: Text(
                    '/ 100',
                    style: TextStyle(
                      fontSize: 16,
                      color: colorScheme.onSurfaceVariant,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 20),
            // Factor bars
            _buildFactorBar('Supply Concentration',
                risk.supplyConcentration, colorScheme),
            const SizedBox(height: 8),
            _buildFactorBar('Geopolitical Tension',
                risk.geopoliticalTension, colorScheme),
            const SizedBox(height: 8),
            _buildFactorBar(
                'Trade Policy Signal', risk.tradePolicySignal, colorScheme),
            const SizedBox(height: 8),
            _buildFactorBar('Logistics Risk', risk.logisticsRisk, colorScheme),
          ],
        ),
      ),
    );
  }

  Widget _buildFactorBar(
      String label, double value, ColorScheme colorScheme) {
    final barColor = _getRiskColor(value);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              label,
              style: TextStyle(
                fontSize: 12,
                color: colorScheme.onSurfaceVariant,
              ),
            ),
            Text(
              '${value.toStringAsFixed(0)}%',
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w600,
                color: barColor,
              ),
            ),
          ],
        ),
        const SizedBox(height: 4),
        ClipRRect(
          borderRadius: BorderRadius.circular(4),
          child: LinearProgressIndicator(
            value: value / 100,
            backgroundColor: colorScheme.surfaceContainerLow,
            valueColor: AlwaysStoppedAnimation(barColor),
            minHeight: 6,
          ),
        ),
      ],
    );
  }

  Widget _buildStatsRow(ColorScheme colorScheme) {
    if (_overview == null) return const SizedBox.shrink();

    return LayoutBuilder(builder: (context, constraints) {
      // Determine if we should stack the cards or keep them side-by-side
      final bool isTight = constraints.maxWidth < 800;

      if (isTight) {
        return Column(
          children: [
            Row(
              children: [
                _buildStatCard(
                  Icons.event_note,
                  '${_overview!.recentEvents}',
                  'Tracked Events',
                  colorScheme.primary,
                  colorScheme,
                ),
                const SizedBox(width: 16),
                _buildStatCard(
                  Icons.warning_amber_rounded,
                  '${_overview!.highRiskZones}',
                  'High-Risk Zones',
                  Colors.red.shade400,
                  colorScheme,
                ),
              ],
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                _buildStatCard(
                  Icons.location_on,
                  '${_chokepoints.length}',
                  'Chokepoints',
                  Colors.orange.shade400,
                  colorScheme,
                ),
                const SizedBox(width: 16),
                _buildStatCard(
                  Icons.swap_horiz,
                  '${_tradeFlows.length}',
                  'Trade Flows',
                  Colors.teal.shade400,
                  colorScheme,
                ),
              ],
            ),
          ],
        );
      }

      return Row(
        children: [
          _buildStatCard(
            Icons.event_note,
            '${_overview!.recentEvents}',
            'Tracked Events',
            colorScheme.primary,
            colorScheme,
          ),
          const SizedBox(width: 16),
          _buildStatCard(
            Icons.warning_amber_rounded,
            '${_overview!.highRiskZones}',
            'High-Risk Zones',
            Colors.red.shade400,
            colorScheme,
          ),
          const SizedBox(width: 16),
          _buildStatCard(
            Icons.location_on,
            '${_chokepoints.length}',
            'Chokepoints',
            Colors.orange.shade400,
            colorScheme,
          ),
          const SizedBox(width: 16),
          _buildStatCard(
            Icons.swap_horiz,
            '${_tradeFlows.length}',
            'Trade Flows',
            Colors.teal.shade400,
            colorScheme,
          ),
        ],
      );
    });
  }

  Widget _buildStatCard(IconData icon, String value, String label,
      Color accentColor, ColorScheme colorScheme) {
    return Expanded(
      child: Card(
        color: colorScheme.surfaceContainerHighest,
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Row(
            children: [
              Container(
                width: 44,
                height: 44,
                decoration: BoxDecoration(
                  color: accentColor.withValues(alpha: 0.15),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(icon, color: accentColor, size: 22),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      value,
                      style: TextStyle(
                        fontSize: 24,
                        fontWeight: FontWeight.bold,
                        color: colorScheme.onSurface,
                      ),
                      overflow: TextOverflow.ellipsis,
                    ),
                    Text(
                      label,
                      style: TextStyle(
                        fontSize: 12,
                        color: colorScheme.onSurfaceVariant,
                      ),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildEventsFeed(ColorScheme colorScheme) {
    return Card(
      color: colorScheme.surfaceContainerHighest,
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.rss_feed, color: colorScheme.primary, size: 20),
                const SizedBox(width: 8),
                Text(
                  'Geopolitical Event Feed',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: colorScheme.onSurface,
                  ),
                ),
                const Spacer(),
                Flexible(
                  child: Text(
                    'Powered by GDELT + NVIDIA NIM',
                    style: TextStyle(
                      fontSize: 11,
                      color: colorScheme.onSurfaceVariant,
                    ),
                    textAlign: TextAlign.end,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            if (_events.isEmpty)
              Center(
                child: Padding(
                  padding: const EdgeInsets.all(24),
                  child: Text(
                    'No events loaded',
                    style: TextStyle(color: colorScheme.onSurfaceVariant),
                  ),
                ),
              )
            else
              ..._events.map((event) => _buildEventTile(event, colorScheme)),
          ],
        ),
      ),
    );
  }

  Widget _buildEventTile(GDELTEvent event, ColorScheme colorScheme) {
    final sentimentColor = event.sentimentLabel == 'escalation'
        ? Colors.red.shade400
        : event.sentimentLabel == 'de-escalation'
            ? Colors.green.shade400
            : Colors.grey;

    final sentimentIcon = event.sentimentLabel == 'escalation'
        ? Icons.trending_up
        : event.sentimentLabel == 'de-escalation'
            ? Icons.trending_down
            : Icons.trending_flat;

    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: colorScheme.surfaceContainerLow,
          borderRadius: BorderRadius.circular(12),
          border: Border(
            left: BorderSide(color: sentimentColor, width: 3),
          ),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(sentimentIcon, color: sentimentColor, size: 18),
                const SizedBox(width: 8),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(
                    color: sentimentColor.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Text(
                    event.sentimentLabel.toUpperCase(),
                    style: TextStyle(
                      color: sentimentColor,
                      fontSize: 10,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                const Spacer(),
                Text(
                  '${event.actor1Country}${event.actor2Country.isNotEmpty ? " → ${event.actor2Country}" : ""}',
                  style: TextStyle(
                    fontSize: 11,
                    color: colorScheme.onSurfaceVariant,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              event.description,
              style: TextStyle(
                fontSize: 13,
                color: colorScheme.onSurface,
                height: 1.4,
              ),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Text(
                  'Goldstein: ${event.goldsteinScale.toStringAsFixed(1)}',
                  style: TextStyle(
                    fontSize: 11,
                    color: colorScheme.onSurfaceVariant,
                  ),
                ),
                const SizedBox(width: 16),
                Text(
                  'Relevance: ${(event.relevance * 100).toStringAsFixed(0)}%',
                  style: TextStyle(
                    fontSize: 11,
                    color: colorScheme.onSurfaceVariant,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildChokepointsList(ColorScheme colorScheme) {
    return Card(
      color: colorScheme.surfaceContainerHighest,
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.warning_amber, color: Colors.orange, size: 20),
                const SizedBox(width: 8),
                Text(
                  'Chokepoints',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: colorScheme.onSurface,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            ..._chokepoints
                .map((cp) => _buildChokepointTile(cp, colorScheme)),
          ],
        ),
      ),
    );
  }

  Widget _buildChokepointTile(Chokepoint cp, ColorScheme colorScheme) {
    final riskColor = cp.riskLevel == 'critical'
        ? Colors.red.shade400
        : cp.riskLevel == 'elevated'
            ? Colors.orange.shade400
            : Colors.green.shade400;

    final typeIcon = cp.type == 'production'
        ? Icons.factory_outlined
        : cp.type == 'shipping'
            ? Icons.directions_boat_outlined
            : Icons.precision_manufacturing_outlined;

    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: colorScheme.surfaceContainerLow,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          children: [
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: riskColor.withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Icon(typeIcon, color: riskColor, size: 20),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    cp.name,
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: colorScheme.onSurface,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    '${cp.country} · ${cp.resource} · ${cp.globalSharePct.toStringAsFixed(0)}% global share',
                    style: TextStyle(
                      fontSize: 11,
                      color: colorScheme.onSurfaceVariant,
                    ),
                  ),
                ],
              ),
            ),
            Container(
              padding:
                  const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: riskColor.withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                cp.riskLevel.toUpperCase(),
                style: TextStyle(
                  color: riskColor,
                  fontSize: 10,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTradeFlowsTable(ColorScheme colorScheme) {
    return Card(
      color: colorScheme.surfaceContainerHighest,
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.swap_horiz, color: Colors.teal, size: 20),
                const SizedBox(width: 8),
                Text(
                  'Trade Flows',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: colorScheme.onSurface,
                  ),
                ),
                const Spacer(),
                Text(
                  'Source: UN Comtrade',
                  style: TextStyle(
                    fontSize: 11,
                    color: colorScheme.onSurfaceVariant,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: DataTable(
                headingRowColor: WidgetStateProperty.all(
                    colorScheme.surfaceContainerLow),
                columnSpacing: 24,
                columns: const [
                  DataColumn(label: Text('Exporter')),
                  DataColumn(label: Text('Importer')),
                  DataColumn(label: Text('Resource')),
                  DataColumn(label: Text('Value (USD)'), numeric: true),
                  DataColumn(label: Text('Weight (kg)'), numeric: true),
                ],
                rows: _tradeFlows.map((flow) {
                  return DataRow(cells: [
                    DataCell(Text(flow.reporterCountry)),
                    DataCell(Text(flow.partnerCountry)),
                    DataCell(
                      Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 8, vertical: 2),
                        decoration: BoxDecoration(
                          color: flow.resource == 'gallium'
                              ? Colors.blue.withValues(alpha: 0.15)
                              : Colors.purple.withValues(alpha: 0.15),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Text(
                          flow.resource,
                          style: TextStyle(
                            color: flow.resource == 'gallium'
                                ? Colors.blue.shade300
                                : Colors.purple.shade300,
                            fontSize: 12,
                          ),
                        ),
                      ),
                    ),
                    DataCell(Text(_formatCurrency(flow.valueUsd))),
                    DataCell(Text(_formatNumber(flow.weightKg))),
                  ]);
                }).toList(),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Color _getRiskColor(double score) {
    if (score >= 70) return Colors.red.shade400;
    if (score >= 40) return Colors.orange.shade400;
    return Colors.green.shade400;
  }

  String _formatCurrency(double value) {
    if (value >= 1000000) {
      return '\$${(value / 1000000).toStringAsFixed(1)}M';
    }
    if (value >= 1000) {
      return '\$${(value / 1000).toStringAsFixed(0)}K';
    }
    return '\$${value.toStringAsFixed(0)}';
  }

  String _formatNumber(double value) {
    if (value >= 1000000) {
      return '${(value / 1000000).toStringAsFixed(1)}M';
    }
    if (value >= 1000) {
      return '${(value / 1000).toStringAsFixed(0)}K';
    }
    return value.toStringAsFixed(0);
  }
}
