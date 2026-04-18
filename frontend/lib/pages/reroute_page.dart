import 'package:flutter/material.dart';
import '../models/models.dart';
import '../services/api_service.dart';

/// Alert Inbox + Manual Simulation page.
class ReroutePage extends StatefulWidget {
  final ApiService apiService;
  const ReroutePage({super.key, required this.apiService});
  @override
  State<ReroutePage> createState() => _ReroutePageState();
}

class _ReroutePageState extends State<ReroutePage> with TickerProviderStateMixin {
  late TabController _tabController;

  // Alert Inbox state
  List<AlertRecord> _alerts = [];
  Map<String, RerouteResult?> _alertReroutes = {};
  String? _expandedAlertId;
  bool _alertsLoading = true;

  // Manual Simulation state
  String _selectedResource = 'gallium';
  List<Resource> _resources = [];
  RerouteResult? _result;
  bool _simLoading = false;
  bool _simulated = false;
  late AnimationController _animController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _animController = AnimationController(vsync: this, duration: const Duration(milliseconds: 800));
    _loadAlerts();
    _loadResources();
  }

  @override
  void dispose() {
    _tabController.dispose();
    _animController.dispose();
    super.dispose();
  }

  Future<void> _loadAlerts() async {
    setState(() => _alertsLoading = true);
    try {
      final alerts = await widget.apiService.getRecentAlerts(limit: 20);
      final reroutes = <String, RerouteResult?>{};
      for (final a in alerts) {
        if (a.resource.isNotEmpty) {
          try {
            reroutes[a.id] = await widget.apiService.getLatestRerouteResult(resource: a.resource);
          } catch (_) {}
        }
      }
      if (mounted) {
        setState(() { _alerts = alerts; _alertReroutes = reroutes; _alertsLoading = false; });
      }
    } catch (e) {
      if (mounted) setState(() => _alertsLoading = false);
    }
  }

  Future<void> _loadResources() async {
    try {
      final res = await widget.apiService.getResources();
      if (mounted) {
        setState(() {
          _resources = res;
          if (_resources.isNotEmpty && !_resources.any((r) => r.id == _selectedResource)) {
            _selectedResource = _resources.first.id;
          }
        });
      }
    } catch (_) {}
  }

  Future<void> _acknowledgeAlert(String id) async {
    try {
      await widget.apiService.acknowledgeAlert(id);
      await _loadAlerts();
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Error: $e')));
    }
  }

  Future<void> _runSimulation() async {
    setState(() { _simLoading = true; _simulated = false; });
    try {
      final result = await widget.apiService.simulateReroute(resource: _selectedResource);
      setState(() { _result = result; _simLoading = false; _simulated = true; });
      _animController.forward(from: 0);
    } catch (e) {
      setState(() => _simLoading = false);
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Simulation error: $e')));
    }
  }

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final unacked = _alerts.where((a) => !a.acknowledged).length;

    return Scaffold(
      backgroundColor: cs.surface,
      body: Column(children: [
        // Header
        Padding(
          padding: const EdgeInsets.fromLTRB(24, 24, 24, 0),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Row(children: [
              Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                Text('Command Center', style: TextStyle(fontSize: 28, fontWeight: FontWeight.bold, color: cs.onSurface)),
                const SizedBox(height: 4),
                Text('Autonomous alerts and supply chain reroute simulations', style: TextStyle(fontSize: 14, color: cs.onSurfaceVariant)),
              ])),
              if (unacked > 0) Container(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                decoration: BoxDecoration(color: Colors.red.shade400.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(20), border: Border.all(color: Colors.red.shade400.withValues(alpha: 0.3))),
                child: Row(mainAxisSize: MainAxisSize.min, children: [
                  Icon(Icons.notifications_active, color: Colors.red.shade400, size: 16),
                  const SizedBox(width: 6),
                  Text('$unacked NEW', style: TextStyle(color: Colors.red.shade400, fontWeight: FontWeight.bold, fontSize: 12)),
                ]),
              ),
            ]),
            const SizedBox(height: 16),
            TabBar(controller: _tabController, tabs: [
              Tab(child: Row(mainAxisSize: MainAxisSize.min, children: [
                const Icon(Icons.inbox_outlined, size: 18), const SizedBox(width: 8), const Text('Alert Inbox'),
                if (unacked > 0) ...[const SizedBox(width: 8), Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(color: Colors.red.shade400, borderRadius: BorderRadius.circular(10)),
                  child: Text('$unacked', style: const TextStyle(color: Colors.white, fontSize: 11, fontWeight: FontWeight.bold)),
                )],
              ])),
              const Tab(child: Row(mainAxisSize: MainAxisSize.min, children: [Icon(Icons.play_circle_outline, size: 18), SizedBox(width: 8), Text('Manual Simulation')])),
            ]),
          ]),
        ),
        // Tab content
        Expanded(child: TabBarView(controller: _tabController, children: [
          _buildAlertInbox(cs),
          _buildManualSimulation(cs),
        ])),
      ]),
    );
  }

  // ══════ ALERT INBOX TAB ══════
  Widget _buildAlertInbox(ColorScheme cs) {
    if (_alertsLoading) return const Center(child: CircularProgressIndicator());
    if (_alerts.isEmpty) return _buildEmptyAlerts(cs);

    return RefreshIndicator(
      onRefresh: _loadAlerts,
      child: ListView.builder(
        padding: const EdgeInsets.all(24),
        itemCount: _alerts.length,
        itemBuilder: (ctx, i) => _buildAlertCard(_alerts[i], cs),
      ),
    );
  }

  Widget _buildEmptyAlerts(ColorScheme cs) {
    return Center(child: Padding(padding: const EdgeInsets.all(48), child: Column(mainAxisSize: MainAxisSize.min, children: [
      Icon(Icons.check_circle_outline, size: 80, color: Colors.green.shade400.withValues(alpha: 0.5)),
      const SizedBox(height: 24),
      Text('All Clear', style: TextStyle(fontSize: 20, fontWeight: FontWeight.w600, color: cs.onSurface)),
      const SizedBox(height: 8),
      Text('No active alerts — your supply chains are\noperating within acceptable risk parameters.', style: TextStyle(fontSize: 14, color: cs.onSurfaceVariant), textAlign: TextAlign.center),
    ])));
  }

  Widget _buildAlertCard(AlertRecord alert, ColorScheme cs) {
    final isCritical = alert.severity == 'critical';
    final sc = isCritical ? Colors.red.shade400 : Colors.orange.shade400;
    final isExpanded = _expandedAlertId == alert.id;
    final reroute = _alertReroutes[alert.id];
    final timeAgo = _formatTimeAgo(alert.createdAt);

    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Card(
        color: cs.surfaceContainerHighest,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16), side: BorderSide(color: alert.acknowledged ? Colors.transparent : sc.withValues(alpha: 0.4), width: alert.acknowledged ? 0 : 1)),
        child: InkWell(
          borderRadius: BorderRadius.circular(16),
          onTap: () => setState(() => _expandedAlertId = isExpanded ? null : alert.id),
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              // Header row
              Row(children: [
                Container(width: 44, height: 44, decoration: BoxDecoration(color: sc.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(12)),
                  child: Stack(children: [
                    Center(child: Icon(isCritical ? Icons.error_outline : Icons.warning_amber, color: sc, size: 24)),
                    if (!alert.acknowledged) Positioned(top: 4, right: 4, child: Container(width: 10, height: 10, decoration: BoxDecoration(color: sc, shape: BoxShape.circle, boxShadow: [BoxShadow(color: sc.withValues(alpha: 0.5), blurRadius: 6)]))),
                  ])),
                const SizedBox(width: 16),
                Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                  Row(children: [
                    Container(padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2), decoration: BoxDecoration(color: sc.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(8)),
                      child: Text(alert.severity.toUpperCase(), style: TextStyle(color: sc, fontSize: 10, fontWeight: FontWeight.bold))),
                    const SizedBox(width: 8),
                    Text(_fmtRes(alert.resource), style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: cs.onSurface)),
                    const Spacer(),
                    Text(timeAgo, style: TextStyle(fontSize: 11, color: cs.onSurfaceVariant)),
                  ]),
                  const SizedBox(height: 4),
                  Text(alert.message, style: TextStyle(fontSize: 13, color: cs.onSurface, height: 1.4), maxLines: isExpanded ? null : 2, overflow: isExpanded ? null : TextOverflow.ellipsis),
                ])),
              ]),
              // Expanded: show reroute result + acknowledge button
              if (isExpanded) ...[
                const SizedBox(height: 16),
                const Divider(height: 1),
                const SizedBox(height: 16),
                // Stats row
                Row(children: [
                  _alertStat('Risk Score', '${alert.riskScore.toStringAsFixed(0)}/100', sc, cs),
                  _alertStat('Threshold', alert.threshold.toStringAsFixed(0), cs.onSurfaceVariant, cs),
                  _alertStat('Alternatives', alert.alternativesCount.toString(), Colors.green.shade400, cs),
                  _alertStat('Region', alert.region, cs.primary, cs),
                ]),
                // Reroute alternatives
                if (reroute != null && reroute.alternatives.isNotEmpty) ...[
                  const SizedBox(height: 16),
                  Text('Autonomous Reroute Results', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: cs.onSurface)),
                  const SizedBox(height: 12),
                  ...reroute.alternatives.asMap().entries.map((e) => _buildAltRow(e.key, e.value, cs)),
                ],
                const SizedBox(height: 16),
                Row(mainAxisAlignment: MainAxisAlignment.end, children: [
                  if (!alert.acknowledged) FilledButton.icon(
                    onPressed: () => _acknowledgeAlert(alert.id),
                    icon: const Icon(Icons.check, size: 18),
                    label: const Text('Acknowledge'),
                  ),
                  if (alert.acknowledged) Chip(
                    avatar: Icon(Icons.check_circle, color: Colors.green.shade400, size: 18),
                    label: Text('Acknowledged', style: TextStyle(color: Colors.green.shade400, fontSize: 12)),
                    backgroundColor: Colors.green.shade400.withValues(alpha: 0.1),
                    side: BorderSide.none,
                  ),
                ]),
              ],
            ]),
          ),
        ),
      ),
    );
  }

  Widget _alertStat(String label, String value, Color color, ColorScheme cs) {
    return Expanded(child: Column(children: [
      Text(value, style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: color)),
      const SizedBox(height: 2),
      Text(label, style: TextStyle(fontSize: 11, color: cs.onSurfaceVariant)),
    ]));
  }

  Widget _buildAltRow(int rank, RerouteAlternative alt, ColorScheme cs) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(color: cs.surfaceContainerLow, borderRadius: BorderRadius.circular(12)),
        child: Row(children: [
          Container(width: 28, height: 28, decoration: BoxDecoration(color: cs.primary.withValues(alpha: 0.15), shape: BoxShape.circle),
            child: Center(child: Text('#${rank + 1}', style: TextStyle(color: cs.primary, fontWeight: FontWeight.bold, fontSize: 12)))),
          const SizedBox(width: 12),
          Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text(alt.supplierName, style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: cs.onSurface)),
            Text(alt.country, style: TextStyle(fontSize: 11, color: cs.onSurfaceVariant)),
          ])),
          Column(crossAxisAlignment: CrossAxisAlignment.end, children: [
            Text('${alt.feasibilityScore.toStringAsFixed(0)}/100', style: TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: _feasColor(alt.feasibilityScore))),
            Text('${alt.leadTimeDays}d lead', style: TextStyle(fontSize: 11, color: cs.onSurfaceVariant)),
          ]),
        ]),
      ),
    );
  }

  // ══════ MANUAL SIMULATION TAB ══════
  Widget _buildManualSimulation(ColorScheme cs) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        _buildSimControls(cs),
        const SizedBox(height: 24),
        if (_simLoading) Center(child: Padding(padding: const EdgeInsets.all(48), child: Column(children: [
          const CircularProgressIndicator(), const SizedBox(height: 24),
          Text('Analyzing disruption scenario...', style: TextStyle(fontSize: 16, color: cs.onSurfaceVariant)),
        ]))),
        if (_simulated && _result != null) _buildSimResults(cs),
        if (_simulated && _result == null) _buildNoDisruption(cs),
        if (!_simulated && !_simLoading) Center(child: Padding(padding: const EdgeInsets.all(48), child: Column(children: [
          Icon(Icons.alt_route, size: 80, color: cs.primary.withValues(alpha: 0.3)), const SizedBox(height: 24),
          Text('Ready to Simulate', style: TextStyle(fontSize: 20, fontWeight: FontWeight.w600, color: cs.onSurface)), const SizedBox(height: 8),
          Text('Select a resource and click "Run Simulation" to test\nhow the supply chain responds to a disruption scenario.', style: TextStyle(fontSize: 14, color: cs.onSurfaceVariant), textAlign: TextAlign.center),
        ]))),
      ]),
    );
  }

  Widget _buildSimControls(ColorScheme cs) {
    return Card(color: cs.surfaceContainerHighest, child: Padding(padding: const EdgeInsets.all(24), child: Row(children: [
      Icon(Icons.alt_route_outlined, color: cs.primary, size: 24), const SizedBox(width: 16),
      Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Text('Disruption Scenario', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: cs.onSurface)),
        Text('Select a resource to run the bypass algorithm', style: TextStyle(fontSize: 12, color: cs.onSurfaceVariant)),
      ])),
      const SizedBox(width: 16),
      if (_resources.isEmpty) const SizedBox(width: 200, child: LinearProgressIndicator(minHeight: 2))
      else SegmentedButton<String>(
        showSelectedIcon: false,
        segments: _resources.map((r) => ButtonSegment(value: r.id, label: Text(r.name), icon: Icon(_resIcon(r.id), size: 18))).toList(),
        selected: {_selectedResource},
        onSelectionChanged: (s) => setState(() { _selectedResource = s.first; _simulated = false; }),
      ),
      const SizedBox(width: 16),
      FilledButton.icon(
        onPressed: _simLoading || _resources.isEmpty ? null : _runSimulation,
        icon: _simLoading ? const SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2)) : const Icon(Icons.play_arrow),
        label: Text(_simLoading ? 'Simulating...' : 'Run Simulation'),
      ),
    ])));
  }

  Widget _buildNoDisruption(ColorScheme cs) {
    return Card(color: Colors.green.shade900.withValues(alpha: 0.3), child: Padding(padding: const EdgeInsets.all(32), child: Row(children: [
      Icon(Icons.check_circle, color: Colors.green.shade400, size: 48), const SizedBox(width: 24),
      Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Text('No Disruption Triggered', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600, color: Colors.green.shade300)),
        const SizedBox(height: 4),
        Text('No regions currently exceed the reroute trigger threshold for $_selectedResource.', style: TextStyle(fontSize: 14, color: cs.onSurface)),
      ])),
    ])));
  }

  Widget _buildSimResults(ColorScheme cs) {
    final r = _result!;
    return FadeTransition(
      opacity: CurvedAnimation(parent: _animController, curve: Curves.easeIn),
      child: Column(children: [
        // Disruption card
        Card(color: Colors.red.shade900.withValues(alpha: 0.3), child: Padding(padding: const EdgeInsets.all(24), child: Row(children: [
          Container(width: 56, height: 56, decoration: BoxDecoration(color: Colors.red.shade400.withValues(alpha: 0.2), borderRadius: BorderRadius.circular(16)),
            child: Icon(Icons.warning_amber, color: Colors.red.shade400, size: 32)),
          const SizedBox(width: 20),
          Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text('⚠ Disruption Detected: ${r.triggerRegion}', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.red.shade300)),
            const SizedBox(height: 4),
            Text('Risk score ${r.triggerRiskScore.toStringAsFixed(0)}/100 exceeds threshold. ${r.alternatives.length} alternative suppliers identified for ${r.resource}.', style: TextStyle(fontSize: 14, color: cs.onSurface, height: 1.4)),
          ])),
        ]))),
        const SizedBox(height: 24),
        Text('Alternative Supply Routes', style: TextStyle(fontSize: 20, fontWeight: FontWeight.w600, color: cs.onSurface)),
        const SizedBox(height: 16),
        Row(crossAxisAlignment: CrossAxisAlignment.start, children: r.alternatives.map((a) => Expanded(child: Padding(padding: const EdgeInsets.symmetric(horizontal: 8), child: _buildAltCard(a, cs)))).toList()),
      ]),
    );
  }

  Widget _buildAltCard(RerouteAlternative alt, ColorScheme cs) {
    return Card(color: cs.surfaceContainerHighest, child: Padding(padding: const EdgeInsets.all(20), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      Row(children: [
        Container(width: 40, height: 40, decoration: BoxDecoration(color: cs.primary.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(10)),
          child: Icon(Icons.factory_outlined, color: cs.primary, size: 20)),
        const SizedBox(width: 12),
        Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(alt.supplierName, style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: cs.onSurface)),
          Text(alt.country, style: TextStyle(fontSize: 12, color: cs.onSurfaceVariant)),
        ])),
      ]),
      const Divider(height: 24),
      _metricRow('Feasibility', '${alt.feasibilityScore.toStringAsFixed(0)}/100', cs),
      const SizedBox(height: 8),
      _metricRow('Capacity', '${alt.capacityTonnes.toStringAsFixed(0)} t/yr', cs),
      const SizedBox(height: 8),
      _metricRow('Absorption', '${alt.absorptionPct.toStringAsFixed(1)}%', cs),
      const SizedBox(height: 8),
      _metricRow('Lead Time', '${alt.leadTimeDays} days', cs),
      const SizedBox(height: 12),
      ClipRRect(borderRadius: BorderRadius.circular(4),
        child: LinearProgressIndicator(value: alt.feasibilityScore / 100, backgroundColor: cs.surfaceContainerLow, valueColor: AlwaysStoppedAnimation(_feasColor(alt.feasibilityScore)), minHeight: 8)),
    ])));
  }

  Widget _metricRow(String l, String v, ColorScheme cs) => Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
    Text(l, style: TextStyle(fontSize: 12, color: cs.onSurfaceVariant)),
    Text(v, style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: cs.onSurface)),
  ]);

  // ── Helpers ──
  Color _feasColor(double s) => s >= 70 ? Colors.green.shade400 : s >= 40 ? Colors.orange.shade400 : Colors.red.shade400;
  IconData _resIcon(String k) => {'gallium': Icons.science_outlined, 'germanium': Icons.memory_outlined, 'lithium': Icons.battery_charging_full_outlined, 'cobalt': Icons.bolt_outlined, 'graphite': Icons.layers_outlined}[k.toLowerCase()] ?? Icons.public_outlined;
  String _fmtRes(String k) => {'gallium': 'Gallium', 'germanium': 'Germanium', 'lithium': 'Lithium', 'cobalt': 'Cobalt', 'graphite': 'Graphite'}[k.toLowerCase()] ?? (k.isEmpty ? '' : k[0].toUpperCase() + k.substring(1));

  String _formatTimeAgo(String isoDate) {
    try {
      final dt = DateTime.parse(isoDate);
      final diff = DateTime.now().difference(dt);
      if (diff.inMinutes < 60) return '${diff.inMinutes}m ago';
      if (diff.inHours < 24) return '${diff.inHours}h ago';
      return '${diff.inDays}d ago';
    } catch (_) { return ''; }
  }
}
