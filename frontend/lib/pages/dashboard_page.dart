import 'package:flutter/material.dart';
import 'package:fl_chart/fl_chart.dart';
import '../models/models.dart';
import '../services/api_service.dart';

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
  Map<String, List<RiskScoreSnapshot>> _riskHistory = {};
  String _selectedChartResource = 'gallium';
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    setState(() { _loading = true; _error = null; });
    try {
      final results = await Future.wait([
        widget.apiService.getRiskOverview(),
        widget.apiService.getRecentEvents(),
        widget.apiService.getChokepoints(),
        widget.apiService.getTradeFlows(),
      ]);
      final overview = results[0] as RiskOverview;

      // Load risk history for all resources
      final historyMap = <String, List<RiskScoreSnapshot>>{};
      for (final key in overview.resourceRisks.keys) {
        try {
          historyMap[key] = await widget.apiService.getRiskHistory(resource: key, days: 30);
        } catch (_) {
          historyMap[key] = [];
        }
      }

      setState(() {
        _overview = overview;
        _events = results[1] as List<GDELTEvent>;
        _chokepoints = results[2] as List<Chokepoint>;
        _tradeFlows = results[3] as List<TradeFlow>;
        _riskHistory = historyMap;
        if (!historyMap.containsKey(_selectedChartResource) && historyMap.isNotEmpty) {
          _selectedChartResource = historyMap.keys.first;
        }
        _loading = false;
      });
    } catch (e) {
      setState(() { _error = e.toString(); _loading = false; });
    }
  }

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Scaffold(
      backgroundColor: cs.surface,
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? _buildErrorState(cs)
              : RefreshIndicator(
                  onRefresh: _loadData,
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.all(24),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        _buildHeader(cs),
                        const SizedBox(height: 24),
                        _buildRiskCards(cs),
                        const SizedBox(height: 24),
                        _buildTrendPanel(cs),
                        const SizedBox(height: 24),
                        _buildStatsRow(cs),
                        const SizedBox(height: 24),
                        Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Expanded(flex: 3, child: _buildEventsFeed(cs)),
                            const SizedBox(width: 24),
                            Expanded(flex: 2, child: _buildChokepointsList(cs)),
                          ],
                        ),
                        const SizedBox(height: 24),
                        _buildTradeFlowsTable(cs),
                      ],
                    ),
                  ),
                ),
    );
  }

  Widget _buildErrorState(ColorScheme cs) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.cloud_off, size: 64, color: cs.error),
          const SizedBox(height: 16),
          Text('Unable to connect to backend', style: TextStyle(fontSize: 20, fontWeight: FontWeight.w600, color: cs.onSurface)),
          const SizedBox(height: 8),
          Text('Make sure the Go server is running on localhost:8080', style: TextStyle(color: cs.onSurfaceVariant)),
          const SizedBox(height: 8),
          Text(_error ?? '', style: TextStyle(color: cs.error, fontSize: 12), textAlign: TextAlign.center),
          const SizedBox(height: 24),
          FilledButton.icon(onPressed: _loadData, icon: const Icon(Icons.refresh), label: const Text('Retry')),
        ],
      ),
    );
  }

  Widget _buildHeader(ColorScheme cs) {
    return Row(
      children: [
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('YanPlatform', style: TextStyle(fontSize: 28, fontWeight: FontWeight.bold, color: cs.onSurface)),
              const SizedBox(height: 4),
              Text(
                _overview == null ? 'Scan ongoing...'
                    : '${_overview!.resourceRisks.keys.map((k) => k[0].toUpperCase() + k.substring(1)).join(', ')} — Real-time Risk Monitoring',
                style: TextStyle(fontSize: 14, color: cs.onSurfaceVariant),
                maxLines: 1, overflow: TextOverflow.ellipsis,
              ),
            ],
          ),
        ),
        const SizedBox(width: 16),
        FilledButton.tonalIcon(onPressed: _loadData, icon: const Icon(Icons.refresh, size: 18), label: const Text('Refresh Data')),
      ],
    );
  }

  // ── Risk Cards with Sparklines ──
  Widget _buildRiskCards(ColorScheme cs) {
    if (_overview == null) return const SizedBox.shrink();
    final risks = _overview!.resourceRisks.entries.toList();
    if (risks.isEmpty) return const SizedBox.shrink();

    final rows = <Widget>[];
    for (var i = 0; i < risks.length; i += 2) {
      final isLastSingle = i + 1 == risks.length;
      rows.add(Row(children: [
        Expanded(child: _buildRiskCard(_fmtRes(risks[i].key), risks[i].value, _resIcon(risks[i].key), cs)),
        const SizedBox(width: 16),
        Expanded(child: isLastSingle ? const SizedBox.shrink()
            : _buildRiskCard(_fmtRes(risks[i + 1].key), risks[i + 1].value, _resIcon(risks[i + 1].key), cs)),
      ]));
      if (i + 2 < risks.length) rows.add(const SizedBox(height: 16));
    }
    return Column(children: rows);
  }

  Widget _buildRiskCard(String title, RiskScore risk, IconData icon, ColorScheme cs) {
    final rc = _riskColor(risk.overallScore);
    final history = _riskHistory[risk.resource] ?? [];

    return Card(
      color: cs.surfaceContainerHighest,
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Row(children: [
            Icon(icon, color: cs.primary, size: 24),
            const SizedBox(width: 12),
            Expanded(child: Text(title, style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600, color: cs.onSurface), overflow: TextOverflow.ellipsis)),
            const SizedBox(width: 8),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
              decoration: BoxDecoration(color: rc.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(20), border: Border.all(color: rc.withValues(alpha: 0.3))),
              child: Text(risk.isHighRisk ? 'HIGH RISK' : 'MODERATE', style: TextStyle(color: rc, fontWeight: FontWeight.bold, fontSize: 12)),
            ),
          ]),
          const SizedBox(height: 20),
          Row(crossAxisAlignment: CrossAxisAlignment.end, children: [
            Text(risk.overallScore.toStringAsFixed(0), style: TextStyle(fontSize: 48, fontWeight: FontWeight.bold, color: rc, height: 1)),
            Padding(padding: const EdgeInsets.only(bottom: 8, left: 4), child: Text('/ 100', style: TextStyle(fontSize: 16, color: cs.onSurfaceVariant))),
            const Spacer(),
            if (history.length >= 2) SizedBox(width: 120, height: 40, child: _buildSparkline(history, rc)),
          ]),
          const SizedBox(height: 20),
          _buildFactorBar('Supply Concentration', risk.supplyConcentration, cs),
          const SizedBox(height: 8),
          _buildFactorBar('Geopolitical Tension', risk.geopoliticalTension, cs),
          const SizedBox(height: 8),
          _buildFactorBar('Trade Policy Signal', risk.tradePolicySignal, cs),
          const SizedBox(height: 8),
          _buildFactorBar('Logistics Risk', risk.logisticsRisk, cs),
        ]),
      ),
    );
  }

  Widget _buildSparkline(List<RiskScoreSnapshot> history, Color color) {
    final spots = history.asMap().entries.map((e) => FlSpot(e.key.toDouble(), e.value.overallScore)).toList();
    return LineChart(
      LineChartData(
        gridData: const FlGridData(show: false),
        titlesData: const FlTitlesData(show: false),
        borderData: FlBorderData(show: false),
        lineTouchData: const LineTouchData(enabled: false),
        lineBarsData: [
          LineChartBarData(
            spots: spots, isCurved: true, color: color, barWidth: 2, dotData: const FlDotData(show: false),
            belowBarData: BarAreaData(show: true, color: color.withValues(alpha: 0.1)),
          ),
        ],
      ),
    );
  }

  Widget _buildFactorBar(String label, double value, ColorScheme cs) {
    final bc = _riskColor(value);
    return Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
        Text(label, style: TextStyle(fontSize: 12, color: cs.onSurfaceVariant)),
        Text('${value.toStringAsFixed(0)}%', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: bc)),
      ]),
      const SizedBox(height: 4),
      ClipRRect(
        borderRadius: BorderRadius.circular(4),
        child: LinearProgressIndicator(value: value / 100, backgroundColor: cs.surfaceContainerLow, valueColor: AlwaysStoppedAnimation(bc), minHeight: 6),
      ),
    ]);
  }

  // ── Full Trend Panel ──
  Widget _buildTrendPanel(ColorScheme cs) {
    final history = _riskHistory[_selectedChartResource] ?? [];
    return Card(
      color: cs.surfaceContainerHighest,
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Row(children: [
            Icon(Icons.show_chart, color: cs.primary, size: 20),
            const SizedBox(width: 8),
            Text('30-Day Risk Trend', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: cs.onSurface)),
            const Spacer(),
            if (_riskHistory.isNotEmpty)
              SegmentedButton<String>(
                showSelectedIcon: false,
                style: SegmentedButton.styleFrom(visualDensity: VisualDensity.compact),
                segments: _riskHistory.keys.map((k) => ButtonSegment(value: k, label: Text(_fmtRes(k).split(' ').first, style: const TextStyle(fontSize: 11)))).toList(),
                selected: {_selectedChartResource},
                onSelectionChanged: (s) => setState(() => _selectedChartResource = s.first),
              ),
          ]),
          const SizedBox(height: 24),
          SizedBox(
            height: 280,
            child: history.length < 2
                ? Center(child: Text('No trend data available', style: TextStyle(color: cs.onSurfaceVariant)))
                : _buildTrendChart(history, cs),
          ),
        ]),
      ),
    );
  }

  Widget _buildTrendChart(List<RiskScoreSnapshot> history, ColorScheme cs) {
    final overallSpots = history.asMap().entries.map((e) => FlSpot(e.key.toDouble(), e.value.overallScore)).toList();
    final tensionSpots = history.asMap().entries.map((e) => FlSpot(e.key.toDouble(), e.value.geopoliticalTension)).toList();
    final policySpots = history.asMap().entries.map((e) => FlSpot(e.key.toDouble(), e.value.tradePolicySignal)).toList();

    return LineChart(
      LineChartData(
        minY: 0, maxY: 100,
        gridData: FlGridData(
          show: true,
          drawVerticalLine: false,
          horizontalInterval: 20,
          getDrawingHorizontalLine: (v) => FlLine(color: cs.outlineVariant.withValues(alpha: 0.3), strokeWidth: 1),
        ),
        titlesData: FlTitlesData(
          leftTitles: AxisTitles(sideTitles: SideTitles(showTitles: true, reservedSize: 40, interval: 20,
            getTitlesWidget: (v, _) => Text('${v.toInt()}', style: TextStyle(fontSize: 10, color: cs.onSurfaceVariant)))),
          bottomTitles: AxisTitles(sideTitles: SideTitles(showTitles: true, reservedSize: 28, interval: (history.length / 6).ceilToDouble().clamp(1, 10),
            getTitlesWidget: (v, _) {
              final i = v.toInt();
              if (i < 0 || i >= history.length) return const SizedBox.shrink();
              final d = history[i].date;
              return Text(d.length >= 10 ? '${d.substring(5, 7)}/${d.substring(8)}' : d, style: TextStyle(fontSize: 9, color: cs.onSurfaceVariant));
            })),
          topTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
          rightTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
        ),
        borderData: FlBorderData(show: true, border: Border(bottom: BorderSide(color: cs.outlineVariant.withValues(alpha: 0.3)), left: BorderSide(color: cs.outlineVariant.withValues(alpha: 0.3)))),
        extraLinesData: ExtraLinesData(horizontalLines: [
          HorizontalLine(y: 70, color: Colors.red.shade400.withValues(alpha: 0.5), strokeWidth: 1, dashArray: [8, 4],
            label: HorizontalLineLabel(show: true, alignment: Alignment.topRight, style: TextStyle(fontSize: 10, color: Colors.red.shade400), labelResolver: (_) => 'Threshold: 70')),
        ]),
        lineTouchData: LineTouchData(
          touchTooltipData: LineTouchTooltipData(
            getTooltipColor: (_) => cs.surfaceContainerHighest,
            getTooltipItems: (spots) => spots.map((s) {
              final labels = ['Overall', 'Tension', 'Policy'];
              final colors = [Colors.amber, Colors.red.shade300, Colors.blue.shade300];
              return LineTooltipItem('${labels[s.barIndex]}: ${s.y.toStringAsFixed(1)}', TextStyle(color: colors[s.barIndex], fontSize: 12, fontWeight: FontWeight.w600));
            }).toList(),
          ),
        ),
        lineBarsData: [
          LineChartBarData(spots: overallSpots, isCurved: true, color: Colors.amber, barWidth: 3, dotData: const FlDotData(show: false),
            belowBarData: BarAreaData(show: true, gradient: LinearGradient(begin: Alignment.topCenter, end: Alignment.bottomCenter, colors: [Colors.amber.withValues(alpha: 0.2), Colors.amber.withValues(alpha: 0.0)]))),
          LineChartBarData(spots: tensionSpots, isCurved: true, color: Colors.red.shade300, barWidth: 1.5, dotData: const FlDotData(show: false), dashArray: [4, 3]),
          LineChartBarData(spots: policySpots, isCurved: true, color: Colors.blue.shade300, barWidth: 1.5, dotData: const FlDotData(show: false), dashArray: [4, 3]),
        ],
      ),
    );
  }

  // ── Stats Row ──
  Widget _buildStatsRow(ColorScheme cs) {
    if (_overview == null) return const SizedBox.shrink();
    return Row(children: [
      _buildStatCard(Icons.event_note, '${_overview!.recentEvents}', 'Tracked Events', cs.primary, cs),
      const SizedBox(width: 16),
      _buildStatCard(Icons.warning_amber_rounded, '${_overview!.highRiskZones}', 'High-Risk Zones', Colors.red.shade400, cs),
      const SizedBox(width: 16),
      _buildStatCard(Icons.location_on, '${_chokepoints.length}', 'Chokepoints', Colors.orange.shade400, cs),
      const SizedBox(width: 16),
      _buildStatCard(Icons.swap_horiz, '${_tradeFlows.length}', 'Trade Flows', Colors.teal.shade400, cs),
    ]);
  }

  Widget _buildStatCard(IconData icon, String value, String label, Color accent, ColorScheme cs) {
    return Expanded(
      child: Card(color: cs.surfaceContainerHighest, child: Padding(padding: const EdgeInsets.all(20), child: Row(children: [
        Container(width: 44, height: 44, decoration: BoxDecoration(color: accent.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(12)),
          child: Icon(icon, color: accent, size: 22)),
        const SizedBox(width: 16),
        Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, mainAxisSize: MainAxisSize.min, children: [
          Text(value, style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: cs.onSurface), overflow: TextOverflow.ellipsis),
          Text(label, style: TextStyle(fontSize: 12, color: cs.onSurfaceVariant), overflow: TextOverflow.ellipsis),
        ])),
      ]))),
    );
  }

  // ── Events Feed ──
  Widget _buildEventsFeed(ColorScheme cs) {
    return Card(
      color: cs.surfaceContainerHighest,
      child: Padding(padding: const EdgeInsets.all(24), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          Icon(Icons.rss_feed, color: cs.primary, size: 20), const SizedBox(width: 8),
          Text('Geopolitical Event Feed', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: cs.onSurface)),
          const Spacer(),
          Flexible(child: Text('Powered by GDELT + NVIDIA NIM', style: TextStyle(fontSize: 11, color: cs.onSurfaceVariant), textAlign: TextAlign.end, overflow: TextOverflow.ellipsis)),
        ]),
        const SizedBox(height: 16),
        if (_events.isEmpty) Center(child: Padding(padding: const EdgeInsets.all(24), child: Text('No events loaded', style: TextStyle(color: cs.onSurfaceVariant))))
        else ..._events.map((e) => _buildEventTile(e, cs)),
      ])),
    );
  }

  Widget _buildEventTile(GDELTEvent event, ColorScheme cs) {
    final sc = event.sentimentLabel == 'escalation' ? Colors.red.shade400 : event.sentimentLabel == 'de-escalation' ? Colors.green.shade400 : Colors.grey;
    final si = event.sentimentLabel == 'escalation' ? Icons.trending_up : event.sentimentLabel == 'de-escalation' ? Icons.trending_down : Icons.trending_flat;
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(color: cs.surfaceContainerLow, borderRadius: BorderRadius.circular(12), border: Border(left: BorderSide(color: sc, width: 3))),
        child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Row(children: [
            Icon(si, color: sc, size: 18), const SizedBox(width: 8),
            Container(padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2), decoration: BoxDecoration(color: sc.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(10)),
              child: Text(event.sentimentLabel.toUpperCase(), style: TextStyle(color: sc, fontSize: 10, fontWeight: FontWeight.bold))),
            const Spacer(),
            Text('${event.actor1Country}${event.actor2Country.isNotEmpty ? " → ${event.actor2Country}" : ""}', style: TextStyle(fontSize: 11, color: cs.onSurfaceVariant)),
          ]),
          const SizedBox(height: 8),
          Text(event.description, style: TextStyle(fontSize: 13, color: cs.onSurface, height: 1.4)),
          const SizedBox(height: 8),
          Row(children: [
            Text('Goldstein: ${event.goldsteinScale.toStringAsFixed(1)}', style: TextStyle(fontSize: 11, color: cs.onSurfaceVariant)),
            const SizedBox(width: 16),
            Text('Relevance: ${(event.relevance * 100).toStringAsFixed(0)}%', style: TextStyle(fontSize: 11, color: cs.onSurfaceVariant)),
          ]),
        ]),
      ),
    );
  }

  // ── Chokepoints ──
  Widget _buildChokepointsList(ColorScheme cs) {
    return Card(
      color: cs.surfaceContainerHighest,
      child: Padding(padding: const EdgeInsets.all(24), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [Icon(Icons.warning_amber, color: Colors.orange, size: 20), const SizedBox(width: 8),
          Text('Chokepoints', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: cs.onSurface))]),
        const SizedBox(height: 16),
        ..._chokepoints.map((cp) => _buildChokepointTile(cp, cs)),
      ])),
    );
  }

  Widget _buildChokepointTile(Chokepoint cp, ColorScheme cs) {
    final rc = cp.riskLevel == 'critical' ? Colors.red.shade400 : cp.riskLevel == 'elevated' ? Colors.orange.shade400 : Colors.green.shade400;
    final ti = cp.type == 'production' ? Icons.factory_outlined : cp.type == 'shipping' ? Icons.directions_boat_outlined : Icons.precision_manufacturing_outlined;
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Container(padding: const EdgeInsets.all(14), decoration: BoxDecoration(color: cs.surfaceContainerLow, borderRadius: BorderRadius.circular(12)),
        child: Row(children: [
          Container(width: 40, height: 40, decoration: BoxDecoration(color: rc.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(10)), child: Icon(ti, color: rc, size: 20)),
          const SizedBox(width: 12),
          Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text(cp.name, style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: cs.onSurface)),
            const SizedBox(height: 2),
            Text('${cp.country} · ${cp.resource} · ${cp.globalSharePct.toStringAsFixed(0)}% global share', style: TextStyle(fontSize: 11, color: cs.onSurfaceVariant)),
          ])),
          Container(padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4), decoration: BoxDecoration(color: rc.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(8)),
            child: Text(cp.riskLevel.toUpperCase(), style: TextStyle(color: rc, fontSize: 10, fontWeight: FontWeight.bold))),
        ])),
    );
  }

  // ── Trade Flows ──
  Widget _buildTradeFlowsTable(ColorScheme cs) {
    return Card(
      color: cs.surfaceContainerHighest,
      child: Padding(padding: const EdgeInsets.all(24), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          Icon(Icons.swap_horiz, color: Colors.teal, size: 20), const SizedBox(width: 8),
          Text('Trade Flows', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: cs.onSurface)),
          const Spacer(),
          Text('Source: UN Comtrade', style: TextStyle(fontSize: 11, color: cs.onSurfaceVariant)),
        ]),
        const SizedBox(height: 16),
        SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          child: DataTable(
            headingRowColor: WidgetStateProperty.all(cs.surfaceContainerLow), columnSpacing: 24,
            columns: const [DataColumn(label: Text('Exporter')), DataColumn(label: Text('Importer')), DataColumn(label: Text('Resource')), DataColumn(label: Text('Value (USD)'), numeric: true), DataColumn(label: Text('Weight (kg)'), numeric: true)],
            rows: _tradeFlows.map((f) => DataRow(cells: [
              DataCell(Text(f.reporterCountry)), DataCell(Text(f.partnerCountry)),
              DataCell(Container(padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(color: f.resource == 'gallium' ? Colors.blue.withValues(alpha: 0.15) : Colors.purple.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(8)),
                child: Text(f.resource, style: TextStyle(color: f.resource == 'gallium' ? Colors.blue.shade300 : Colors.purple.shade300, fontSize: 12)))),
              DataCell(Text(_fmtCur(f.valueUsd))), DataCell(Text(_fmtNum(f.weightKg))),
            ])).toList(),
          ),
        ),
      ])),
    );
  }

  // ── Helpers ──
  Color _riskColor(double s) => s >= 70 ? Colors.red.shade400 : s >= 40 ? Colors.orange.shade400 : Colors.green.shade400;
  String _fmtCur(double v) => v >= 1e6 ? '\$${(v / 1e6).toStringAsFixed(1)}M' : v >= 1e3 ? '\$${(v / 1e3).toStringAsFixed(0)}K' : '\$${v.toStringAsFixed(0)}';
  String _fmtNum(double v) => v >= 1e6 ? '${(v / 1e6).toStringAsFixed(1)}M' : v >= 1e3 ? '${(v / 1e3).toStringAsFixed(0)}K' : v.toStringAsFixed(0);
  String _fmtRes(String k) => {'gallium': 'Gallium (Ga)', 'germanium': 'Germanium (Ge)', 'lithium': 'Lithium (Li)', 'cobalt': 'Cobalt (Co)', 'graphite': 'Graphite (C)'}[k.toLowerCase()] ?? (k.isEmpty ? '' : k[0].toUpperCase() + k.substring(1));
  IconData _resIcon(String k) => {'gallium': Icons.science_outlined, 'germanium': Icons.memory_outlined, 'lithium': Icons.battery_charging_full_outlined, 'cobalt': Icons.bolt_outlined, 'graphite': Icons.layers_outlined}[k.toLowerCase()] ?? Icons.public_outlined;
}
