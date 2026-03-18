import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/models.dart';

/// Service for communicating with the Go backend REST API.
class ApiService {
  final String baseUrl;
  final http.Client _client;

  ApiService({this.baseUrl = 'http://localhost:8080'})
      : _client = http.Client();

  /// Fetch the risk overview (gallium + germanium summary).
  Future<RiskOverview> getRiskOverview() async {
    final response = await _client.get(Uri.parse('$baseUrl/api/risk/overview'));
    _checkResponse(response);
    return RiskOverview.fromJson(json.decode(response.body));
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

  /// Run a shadow reroute simulation.
  Future<RerouteResult?> simulateReroute({String resource = 'gallium'}) async {
    final response = await _client
        .get(Uri.parse('$baseUrl/api/reroute/simulate?resource=$resource'));
    _checkResponse(response);
    final data = json.decode(response.body);
    if (data['status'] == 'no_disruption') return null;
    return RerouteResult.fromJson(data);
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
