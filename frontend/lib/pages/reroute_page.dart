import 'package:flutter/material.dart';
import '../models/models.dart';
import '../services/api_service.dart';

/// Shadow Reroute simulation page.
class ReroutePage extends StatefulWidget {
  final ApiService apiService;

  const ReroutePage({super.key, required this.apiService});

  @override
  State<ReroutePage> createState() => _ReroutePageState();
}

class _ReroutePageState extends State<ReroutePage>
    with SingleTickerProviderStateMixin {
  String _selectedResource = 'gallium';
  List<Resource> _resources = [];
  RerouteResult? _result;
  bool _loading = false;
  bool _simulated = false;
  late AnimationController _animController;

  @override
  void initState() {
    super.initState();
    _animController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 800),
    );
    _loadResources();
  }

  Future<void> _loadResources() async {
    try {
      final res = await widget.apiService.getResources();
      if (mounted) {
        setState(() {
          _resources = res;
          if (_resources.isNotEmpty &&
              !_resources.any((r) => r.id == _selectedResource)) {
            _selectedResource = _resources.first.id;
          }
        });
      }
    } catch (e) {
      debugPrint('Error loading resources: $e');
    }
  }

  @override
  void dispose() {
    _animController.dispose();
    super.dispose();
  }

  Future<void> _runSimulation() async {
    setState(() {
      _loading = true;
      _simulated = false;
    });

    try {
      final result = await widget.apiService
          .simulateReroute(resource: _selectedResource);
      setState(() {
        _result = result;
        _loading = false;
        _simulated = true;
      });
      _animController.forward(from: 0);
    } catch (e) {
      setState(() {
        _loading = false;
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Simulation error: $e')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Scaffold(
      backgroundColor: colorScheme.surface,
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildHeader(colorScheme),
            const SizedBox(height: 24),
            _buildSimulationControls(colorScheme),
            const SizedBox(height: 24),
            if (_loading) _buildLoadingState(colorScheme),
            if (_simulated && _result != null) _buildResults(colorScheme),
            if (_simulated && _result == null) _buildNoDisruption(colorScheme),
            if (!_simulated && !_loading) _buildInitialState(colorScheme),
          ],
        ),
      ),
    );
  }

  Widget _buildHeader(ColorScheme colorScheme) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Shadow Reroute Simulation',
          style: TextStyle(
            fontSize: 28,
            fontWeight: FontWeight.bold,
            color: colorScheme.onSurface,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          'Simulate supply chain disruptions and identify alternative supply routes',
          style: TextStyle(
            fontSize: 14,
            color: colorScheme.onSurfaceVariant,
          ),
        ),
      ],
    );
  }

  Widget _buildSimulationControls(ColorScheme colorScheme) {
    return Card(
      color: colorScheme.surfaceContainerHighest,
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Row(
          children: [
            Icon(Icons.alt_route_outlined,
                color: colorScheme.primary, size: 24),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Disruption Scenario',
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w600,
                      color: colorScheme.onSurface,
                    ),
                  ),
                  Text(
                    'Select a resource to run the bypass algorithm',
                    style: TextStyle(
                      fontSize: 12,
                      color: colorScheme.onSurfaceVariant,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 16),
            if (_resources.isEmpty)
              const SizedBox(
                  width: 200, child: LinearProgressIndicator(minHeight: 2))
            else
              SegmentedButton<String>(
                showSelectedIcon: false,
                segments: _resources.map((r) {
                  return ButtonSegment(
                    value: r.id,
                    label: Text(r.name),
                    icon: Icon(_getResourceIcon(r.id), size: 18),
                  );
                }).toList(),
                selected: {_selectedResource},
                onSelectionChanged: (set) {
                  setState(() {
                    _selectedResource = set.first;
                    _simulated = false;
                  });
                },
              ),
            const SizedBox(width: 16),
            FilledButton.icon(
              onPressed: _loading || _resources.isEmpty ? null : _runSimulation,
              icon: _loading
                  ? const SizedBox(
                      width: 18,
                      height: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.play_arrow),
              label: Text(_loading ? 'Simulating...' : 'Run Simulation'),
            ),
          ],
        ),
      ),
    );
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

  Widget _buildLoadingState(ColorScheme colorScheme) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(48),
        child: Column(
          children: [
            const CircularProgressIndicator(),
            const SizedBox(height: 24),
            Text(
              'Analyzing disruption scenario...',
              style: TextStyle(
                fontSize: 16,
                color: colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'Identifying alternative suppliers in politically neutral zones',
              style: TextStyle(
                fontSize: 13,
                color: colorScheme.onSurfaceVariant,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildInitialState(ColorScheme colorScheme) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(48),
        child: Column(
          children: [
            Icon(Icons.alt_route,
                size: 80, color: colorScheme.primary.withValues(alpha: 0.3)),
            const SizedBox(height: 24),
            Text(
              'Ready to Simulate',
              style: TextStyle(
                fontSize: 20,
                fontWeight: FontWeight.w600,
                color: colorScheme.onSurface,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'Select a resource and click "Run Simulation" to test\nhow the supply chain responds to a disruption scenario.',
              style: TextStyle(
                fontSize: 14,
                color: colorScheme.onSurfaceVariant,
              ),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildNoDisruption(ColorScheme colorScheme) {
    return Card(
      color: Colors.green.shade900.withValues(alpha: 0.3),
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Row(
          children: [
            Icon(Icons.check_circle, color: Colors.green.shade400, size: 48),
            const SizedBox(width: 24),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'No Disruption Triggered',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.w600,
                      color: Colors.green.shade300,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    'No regions currently exceed the reroute trigger threshold for $_selectedResource. Supply chain risk is within acceptable bounds.',
                    style: TextStyle(
                      fontSize: 14,
                      color: colorScheme.onSurface,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildResults(ColorScheme colorScheme) {
    final result = _result!;

    return FadeTransition(
      opacity: CurvedAnimation(parent: _animController, curve: Curves.easeIn),
      child: Column(
        children: [
          // Disruption trigger card
          _buildDisruptionCard(result, colorScheme),
          const SizedBox(height: 24),
          // Alternative suppliers
          Text(
            'Alternative Supply Routes',
            style: TextStyle(
              fontSize: 20,
              fontWeight: FontWeight.w600,
              color: colorScheme.onSurface,
            ),
          ),
          const SizedBox(height: 16),
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: result.alternatives
                .map((alt) => Expanded(
                      child: Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 8),
                        child: _buildAlternativeCard(alt, colorScheme),
                      ),
                    ))
                .toList(),
          ),
          const SizedBox(height: 24),
          // Comparison table
          _buildComparisonTable(result, colorScheme),
        ],
      ),
    );
  }

  Widget _buildDisruptionCard(
      RerouteResult result, ColorScheme colorScheme) {
    return Card(
      color: Colors.red.shade900.withValues(alpha: 0.3),
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Row(
          children: [
            Container(
              width: 56,
              height: 56,
              decoration: BoxDecoration(
                color: Colors.red.shade400.withValues(alpha: 0.2),
                borderRadius: BorderRadius.circular(16),
              ),
              child: Icon(Icons.warning_amber,
                  color: Colors.red.shade400, size: 32),
            ),
            const SizedBox(width: 20),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '⚠ Disruption Detected: ${result.triggerRegion}',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: Colors.red.shade300,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    'Risk score ${result.triggerRiskScore.toStringAsFixed(0)}/100 exceeds reroute threshold. '
                    'The system has identified ${result.alternatives.length} alternative suppliers '
                    'in politically neutral zones that could absorb disrupted ${result.resource} supply.',
                    style: TextStyle(
                      fontSize: 14,
                      color: colorScheme.onSurface,
                      height: 1.4,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildAlternativeCard(
      RerouteAlternative alt, ColorScheme colorScheme) {
    return Card(
      color: colorScheme.surfaceContainerHighest,
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(
                    color: colorScheme.primary.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Icon(Icons.factory_outlined,
                      color: colorScheme.primary, size: 20),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        alt.supplierName,
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                          color: colorScheme.onSurface,
                        ),
                      ),
                      Text(
                        alt.country,
                        style: TextStyle(
                          fontSize: 12,
                          color: colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const Divider(height: 24),
            _buildMetricRow('Feasibility',
                '${alt.feasibilityScore.toStringAsFixed(0)}/100', colorScheme),
            const SizedBox(height: 8),
            _buildMetricRow(
                'Capacity', '${alt.capacityTonnes.toStringAsFixed(0)} t/yr',
                colorScheme),
            const SizedBox(height: 8),
            _buildMetricRow('Demand Absorption',
                '${alt.absorptionPct.toStringAsFixed(1)}%', colorScheme),
            const SizedBox(height: 8),
            _buildMetricRow(
                'Lead Time', '${alt.leadTimeDays} days', colorScheme),
            const SizedBox(height: 12),
            // Feasibility bar
            ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: LinearProgressIndicator(
                value: alt.feasibilityScore / 100,
                backgroundColor: colorScheme.surfaceContainerLow,
                valueColor:
                    AlwaysStoppedAnimation(_getRiskColor(alt.feasibilityScore)),
                minHeight: 8,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildMetricRow(
      String label, String value, ColorScheme colorScheme) {
    return Row(
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
          value,
          style: TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.w600,
            color: colorScheme.onSurface,
          ),
        ),
      ],
    );
  }

  Widget _buildComparisonTable(
      RerouteResult result, ColorScheme colorScheme) {
    return Card(
      color: colorScheme.surfaceContainerHighest,
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Supplier Comparison',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: colorScheme.onSurface,
              ),
            ),
            const SizedBox(height: 16),
            SizedBox(
              width: double.infinity,
              child: DataTable(
                headingRowColor: WidgetStateProperty.all(
                    colorScheme.surfaceContainerLow),
                columns: const [
                  DataColumn(label: Text('Rank')),
                  DataColumn(label: Text('Supplier')),
                  DataColumn(label: Text('Country')),
                  DataColumn(
                      label: Text('Capacity (t/yr)'), numeric: true),
                  DataColumn(
                      label: Text('Absorption %'), numeric: true),
                  DataColumn(
                      label: Text('Feasibility'), numeric: true),
                  DataColumn(
                      label: Text('Lead Time'), numeric: true),
                ],
                rows: result.alternatives.asMap().entries.map((entry) {
                  final i = entry.key;
                  final alt = entry.value;
                  return DataRow(
                    cells: [
                      DataCell(
                        Container(
                          width: 28,
                          height: 28,
                          decoration: BoxDecoration(
                            color:
                                colorScheme.primary.withValues(alpha: 0.15),
                            shape: BoxShape.circle,
                          ),
                          child: Center(
                            child: Text(
                              '#${i + 1}',
                              style: TextStyle(
                                color: colorScheme.primary,
                                fontWeight: FontWeight.bold,
                                fontSize: 12,
                              ),
                            ),
                          ),
                        ),
                      ),
                      DataCell(Text(alt.supplierName)),
                      DataCell(Text(alt.country)),
                      DataCell(
                          Text(alt.capacityTonnes.toStringAsFixed(0))),
                      DataCell(
                          Text('${alt.absorptionPct.toStringAsFixed(1)}%')),
                      DataCell(
                        Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Text(alt.feasibilityScore
                                .toStringAsFixed(0)),
                            const SizedBox(width: 4),
                            Icon(
                              Icons.circle,
                              size: 8,
                              color: _getRiskColor(alt.feasibilityScore),
                            ),
                          ],
                        ),
                      ),
                      DataCell(Text('${alt.leadTimeDays} days')),
                    ],
                  );
                }).toList(),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Color _getRiskColor(double score) {
    if (score >= 70) return Colors.green.shade400;
    if (score >= 40) return Colors.orange.shade400;
    return Colors.red.shade400;
  }
}
