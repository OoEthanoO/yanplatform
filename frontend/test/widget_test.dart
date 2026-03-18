import 'package:flutter_test/flutter_test.dart';
import 'package:yanplatform/main.dart';

void main() {
  testWidgets('App renders without errors', (WidgetTester tester) async {
    await tester.pumpWidget(const YanPlatformApp());
    expect(find.text('Dashboard'), findsOneWidget);
  });
}
