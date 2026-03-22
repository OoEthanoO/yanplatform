import re

with open("internal/store/store.go", "r") as f:
    content = f.read()

# Replace Store struct and New func with the Interface and MemoryStore
top_chunk = """// Store defines the data access interface for YanPlatform.
type Store interface {
	LoadSupplierSeed(path string) error
	GetSuppliers(resource string) ([]models.Supplier, error)
	GetAlternativeSuppliers(resource string) ([]models.Supplier, error)
	SaveRiskScore(score models.RiskScore) error
	GetRiskScores(resource string) ([]models.RiskScore, error)
	GetHighRiskZones(threshold float64) ([]models.RiskScore, error)
	SaveEvent(event models.GDELTEvent) error
	GetRecentEvents(limit int) ([]models.GDELTEvent, error)
	SaveTradeFlow(flow models.TradeFlow) error
	GetTradeFlows(resource string) ([]models.TradeFlow, error)
	SaveRerouteResult(result models.RerouteResult) error
	GetLatestRerouteResult(resource string) (*models.RerouteResult, error)
	SaveChokepoint(cp models.Chokepoint) error
	GetChokepoints(resource string) ([]models.Chokepoint, error)
	SeedInitialData() error
}

// MemoryStore provides an in-memory implementation of Store.
type MemoryStore struct {
	mu             sync.RWMutex
	suppliers      []models.Supplier
	riskScores     []models.RiskScore
	events         []models.GDELTEvent
	tradeFlows     []models.TradeFlow
	rerouteResults []models.RerouteResult
	chokepoints    []models.Chokepoint
}

// NewMemoryStore creates a new in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}"""

# Using regex to replace the top section
content = re.sub(
    r"// Store provides data access for all collections\.\ntype Store struct \{.*?\n\}\n\n// New creates a new in-memory store\.\nfunc New\(\) \*Store \{\n\treturn &Store\{\}\n\}",
    top_chunk,
    content,
    flags=re.DOTALL
)

# Change receiver: (s *Store) -> (s *MemoryStore)
content = content.replace("(s *Store)", "(s *MemoryStore)")

# Change func signatures and returns for getters (return array, error or pointer, error)
# For all getters returning slice or pointer:
signatures = [
    (r"func \(s \*MemoryStore\) GetSuppliers\(resource string\) \[\]models\.Supplier \{", r"func (s *MemoryStore) GetSuppliers(resource string) ([]models.Supplier, error) {"),
    (r"func \(s \*MemoryStore\) GetAlternativeSuppliers\(resource string\) \[\]models\.Supplier \{", r"func (s *MemoryStore) GetAlternativeSuppliers(resource string) ([]models.Supplier, error) {"),
    (r"func \(s \*MemoryStore\) GetRiskScores\(resource string\) \[\]models\.RiskScore \{", r"func (s *MemoryStore) GetRiskScores(resource string) ([]models.RiskScore, error) {"),
    (r"func \(s \*MemoryStore\) GetHighRiskZones\(threshold float64\) \[\]models\.RiskScore \{", r"func (s *MemoryStore) GetHighRiskZones(threshold float64) ([]models.RiskScore, error) {"),
    (r"func \(s \*MemoryStore\) GetRecentEvents\(limit int\) \[\]models\.GDELTEvent \{", r"func (s *MemoryStore) GetRecentEvents(limit int) ([]models.GDELTEvent, error) {"),
    (r"func \(s \*MemoryStore\) GetTradeFlows\(resource string\) \[\]models\.TradeFlow \{", r"func (s *MemoryStore) GetTradeFlows(resource string) ([]models.TradeFlow, error) {"),
    (r"func \(s \*MemoryStore\) GetLatestRerouteResult\(resource string\) \*models\.RerouteResult \{", r"func (s *MemoryStore) GetLatestRerouteResult(resource string) (*models.RerouteResult, error) {"),
    (r"func \(s \*MemoryStore\) GetChokepoints\(resource string\) \[\]models\.Chokepoint \{", r"func (s *MemoryStore) GetChokepoints(resource string) ([]models.Chokepoint, error) {")
]

for old, new in signatures:
    content = re.sub(old, new, content)

# Update returns in those getters to append `, nil`
getters = ["GetSuppliers", "GetAlternativeSuppliers", "GetRiskScores", "GetHighRiskZones", "GetRecentEvents", "GetTradeFlows", "GetLatestRerouteResult", "GetChokepoints"]

for match in re.finditer(r"func \(s \*MemoryStore\) (\w+)\(.*?\)(?: \[\]models\.\w+| \*models\.RerouteResult)?, error\) \{.*?\n\}", content, flags=re.DOTALL):
    func_name = match.group(1)
    if func_name in getters:
        body = match.group(0)
        # replace `return <expr>` with `return <expr>, nil`
        # careful not to match func def line
        new_body = re.sub(r"(\n\t+)return (.+?)\n", r"\1return \2, nil\n", body)
        content = content.replace(body, new_body)

# Change signatures for setters
setters = [
    (r"func \(s \*MemoryStore\) SaveRiskScore\(score models\.RiskScore\) \{", r"func (s *MemoryStore) SaveRiskScore(score models.RiskScore) error {"),
    (r"func \(s \*MemoryStore\) SaveEvent\(event models\.GDELTEvent\) \{", r"func (s *MemoryStore) SaveEvent(event models.GDELTEvent) error {"),
    (r"func \(s \*MemoryStore\) SaveTradeFlow\(flow models\.TradeFlow\) \{", r"func (s *MemoryStore) SaveTradeFlow(flow models.TradeFlow) error {"),
    (r"func \(s \*MemoryStore\) SaveRerouteResult\(result models\.RerouteResult\) \{", r"func (s *MemoryStore) SaveRerouteResult(result models.RerouteResult) error {"),
    (r"func \(s \*MemoryStore\) SaveChokepoint\(cp models\.Chokepoint\) \{", r"func (s *MemoryStore) SaveChokepoint(cp models.Chokepoint) error {"),
    (r"func \(s \*MemoryStore\) SeedInitialData\(\) \{", r"func (s *MemoryStore) SeedInitialData() error {")
]
for old, new in setters:
    content = re.sub(old, new, content)

# update setters to return nil at end, and replace empty `return` with `return nil`
for match in re.finditer(r"func \(s \*MemoryStore\) (\w+)\(.*?\)(?: error)? \{.*?\n\}", content, flags=re.DOTALL):
    func_name = match.group(1)
    if any(func_name == setter[0].split()[-1].split('(')[0] for setter in setters):
        body = match.group(0)
        new_body = re.sub(r"(\n\t+)return\n", r"\1return nil\n", body)
        # if the last line of function body before } is not return something
        if not re.search(r"return.*?\n\}$", new_body):
            new_body = re.sub(r"\n\}$", r"\n\treturn nil\n}", new_body)
        content = content.replace(body, new_body)


with open("internal/store/store_refactored.go", "w") as f:
    f.write(content)
