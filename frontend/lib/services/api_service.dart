import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/models.dart';

/// Service for communicating with the Go backend REST API.
class ApiService {
  final String baseUrl;
  final http.Client _client;

  ApiService({this.baseUrl = 'http://localhost:8080'})
      : _client = http.Client();

  /// Fetch the risk overview (dynamic resource summary).
  Future<RiskOverview> getRiskOverview() async {
    final response = await _client.get(Uri.parse('$baseUrl/api/risk/overview'));
    _checkResponse(response);
    return RiskOverview.fromJson(json.decode(response.body));
  }

  /// Fetch all tracked resources.
  Future<List<Resource>> getResources() async {
    final response = await _client.get(Uri.parse('$baseUrl/api/resources'));
    _checkResponse(response);
    final list = json.decode(response.body) as List<dynamic>;
    return list.map((e) => Resource.fromJson(e)).toList();
  }

  /// Fetch chokepoints, optionally filtered by resource.
  Future<List<Chokepoint>> getChokepoints({String? resource}) async {
    final uri = resource != null
        ? Uri.parse('$baseUrl/api/risk/chokepoints?resource=$resource')
        : Uri.parse('$baseUrl/api/risk/chokepoints');
    final response = await _client.get(uri);
    _checkResponse(response);
    final list = json.decode(response.body) as List<dynamic>;
    return list.map((e) => Chokepoint.fromJson(e)).toList();
  }

  /// Fetch risk score trends.
  Future<List<RiskScore>> getRiskTrends() async {
    final response = await _client.get(Uri.parse('$baseUrl/api/risk/trends'));
    _checkResponse(response);
    final list = json.decode(response.body) as List<dynamic>;
    return list.map((e) => RiskScore.fromJson(e)).toList();
  }

  /// Fetch time-series risk history for a resource.
  Future<List<RiskScoreSnapshot>> getRiskHistory({
    required String resource,
    int days = 30,
  }) async {
    final response = await _client.get(
      Uri.parse('$baseUrl/api/risk/history?resource=$resource&days=$days'),
    );
    _checkResponse(response);
    final list = json.decode(response.body) as List<dynamic>?;
    if (list == null) return [];
    return list.map((e) => RiskScoreSnapshot.fromJson(e)).toList();
  }

  /// Run a shadow reroute simulation (on-demand).
  Future<RerouteResult?> simulateReroute({String resource = 'gallium'}) async {
    final response = await _client
        .get(Uri.parse('$baseUrl/api/reroute/simulate?resource=$resource'));
    _checkResponse(response);
    final data = json.decode(response.body);
    if (data['status'] == 'no_disruption') return null;
    return RerouteResult.fromJson(data);
  }

  /// Fetch the latest autonomous reroute result for a resource.
  Future<RerouteResult?> getLatestRerouteResult({
    required String resource,
  }) async {
    final response = await _client.get(
      Uri.parse('$baseUrl/api/reroute/latest?resource=$resource'),
    );
    _checkResponse(response);
    final data = json.decode(response.body);
    if (data['status'] == 'no_results') return null;
    return RerouteResult.fromJson(data);
  }

  /// Fetch reroute simulation history.
  Future<List<RerouteResult>> getRerouteHistory({
    String? resource,
    int limit = 10,
  }) async {
    final query = resource != null
        ? 'resource=$resource&limit=$limit'
        : 'limit=$limit';
    final response = await _client.get(
      Uri.parse('$baseUrl/api/reroute/history?$query'),
    );
    _checkResponse(response);
    final list = json.decode(response.body) as List<dynamic>?;
    if (list == null) return [];
    return list.map((e) => RerouteResult.fromJson(e)).toList();
  }

  /// Fetch recent system alerts.
  Future<List<AlertRecord>> getRecentAlerts({int limit = 20}) async {
    final response = await _client.get(
      Uri.parse('$baseUrl/api/alerts/recent?limit=$limit'),
    );
    _checkResponse(response);
    final list = json.decode(response.body) as List<dynamic>?;
    if (list == null) return [];
    return list.map((e) => AlertRecord.fromJson(e)).toList();
  }

  /// Acknowledge an alert.
  Future<void> acknowledgeAlert(String id) async {
    final response = await _client.post(
      Uri.parse('$baseUrl/api/alerts/$id/acknowledge'),
    );
    _checkResponse(response);
  }

  /// Fetch recent geopolitical events.
  Future<List<GDELTEvent>> getRecentEvents() async {
    final response =
        await _client.get(Uri.parse('$baseUrl/api/events/recent'));
    _checkResponse(response);
    final list = json.decode(response.body) as List<dynamic>;
    return list.map((e) => GDELTEvent.fromJson(e)).toList();
  }

  /// Fetch trade flow data, optionally filtered by resource.
  Future<List<TradeFlow>> getTradeFlows({String? resource}) async {
    final uri = resource != null
        ? Uri.parse('$baseUrl/api/trade/flows?resource=$resource')
        : Uri.parse('$baseUrl/api/trade/flows');
    final response = await _client.get(uri);
    _checkResponse(response);
    final list = json.decode(response.body) as List<dynamic>;
    return list.map((e) => TradeFlow.fromJson(e)).toList();
  }

  /// Fetch supplier list, optionally filtered by resource.
  Future<List<Supplier>> getSuppliers({String? resource}) async {
    final uri = resource != null
        ? Uri.parse('$baseUrl/api/suppliers?resource=$resource')
        : Uri.parse('$baseUrl/api/suppliers');
    final response = await _client.get(uri);
    _checkResponse(response);
    final list = json.decode(response.body) as List<dynamic>;
    return list.map((e) => Supplier.fromJson(e)).toList();
  }

  /// Health check.
  Future<bool> healthCheck() async {
    try {
      final response =
          await _client.get(Uri.parse('$baseUrl/api/health'));
      return response.statusCode == 200;
    } catch (e) {
      return false;
    }
  }

  void _checkResponse(http.Response response) {
    if (response.statusCode != 200) {
      throw Exception(
          'API error ${response.statusCode}: ${response.body}');
    }
  }

  void dispose() {
    _client.close();
  }
}
