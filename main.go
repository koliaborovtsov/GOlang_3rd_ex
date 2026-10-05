package main
import(
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

type Transaction struct{
	ID       string  `json:"id"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
}

type Budget struct{
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	Period   string  `json:"period,omitempty"`
}

type Ledger struct{
	transactions []Transaction
	budgets      map[string]Budget
}

func NewLedger() *Ledger{
	return &Ledger{
		transactions: make([]Transaction, 0),
		budgets:      make(map[string]Budget),
	}
}

func (l *Ledger) SetBudget(b Budget){
	l.budgets[b.Category] = b
}

func (l *Ledger) AddTransaction(tx Transaction) error{
	if tx.Category == ""{
		return errors.New("category is empty")
	}
	if tx.Amount < 0 {
		return errors.New("amount must be non-negative")
	}

	if b, ok := l.budgets[tx.Category]; ok{
		var current float64
		for _, existing := range l.transactions{
			if existing.Category == tx.Category{
				current += existing.Amount
			}
		}

		if current+tx.Amount > b.Limit{
			return errors.New("budget exceeded")
		}
	}

	l.transactions = append(l.transactions, tx)
	return nil
}

func (l *Ledger) Transactions() []Transaction{
	return l.transactions
}

func (l *Ledger) LoadBudgets(r io.Reader) error{
	var budgets []Budget

	if err := json.NewDecoder(r).Decode(&budgets); err != nil{
		return fmt.Errorf("decode budgets JSON: %w", err)
	}

	for i, b := range budgets{
		if b.Category == ""{
			return fmt.Errorf("budget at index %d: empty category", i)
		}
		if b.Limit < 0{
			return fmt.Errorf("budget for category %q: limit must be non-negative", b.Category)
		}
		l.SetBudget(b)
	}

	return nil
}

func ensureSampleBudgetsFile(){
	const filename = "budgets.json"

	if _, err := os.Stat(filename); err == nil{
		return
	} else if !os.IsNotExist(err) {
		log.Printf("cannot check %s: %v", filename, err)
		return
	}

	data := []byte(`[
  {"category":"еда","limit":6000,"period":"месяц"},
  {"category":"транспорт","limit":2500,"period":"месяц"}]`)

	if err := os.WriteFile(filename, data, 0644); err != nil{
		log.Printf("cannot create %s: %v", filename, err)
	} else {
		fmt.Println("Создан пример budgets.json")
	}
}

func main(){
	ledger := NewLedger()

	// Начальные бюджеты
	ledger.SetBudget(Budget{Category: "еда", Limit: 5000, Period: "месяц"})
	ledger.SetBudget(Budget{Category: "транспорт", Limit: 2000, Period: "месяц"})

	fmt.Println("В пределах бюджета")
	tx1 := Transaction{ID: "1", Category: "еда", Amount: 1500}
	if err := ledger.AddTransaction(tx1); err != nil{
		fmt.Printf("Отказ: %v\n", err)
	} else {
		fmt.Printf("Транзакция добавлена: %+v\n", tx1)
	}

	fmt.Println("\nПревышает бюджет")
	tx2 := Transaction{ID: "2", Category: "еда", Amount: 4000} // 1500 + 4000 >5000
	if err := ledger.AddTransaction(tx2); err != nil{
		fmt.Printf("Отказ: %v (транзакция не добавлена)\n", err)
	} else {
		fmt.Printf("Транзакция добавлена: %+v\n", tx2)
	}

	fmt.Println("\nРовно по лимиту")
	tx3 := Transaction{ID: "3", Category: "еда", Amount: 3500} // 1500 + 3500 = 5000
	if err := ledger.AddTransaction(tx3); err != nil{
		fmt.Printf("Отказ: %v\n", err)
	} else {
		fmt.Printf("Транзакция добавлена: %+v\n", tx3)
	}

	fmt.Println("\nКатегория без бюджета")
	tx4 := Transaction{ID: "4", Category: "развлечения", Amount: 100000}
	if err := ledger.AddTransaction(tx4); err != nil{
		fmt.Printf("Отказ: %v\n", err)
	} else {
		fmt.Printf("Транзакция добавлена без проверки бюджета: %+v\n", tx4)
	}

	fmt.Println("\nТекущие транзакции:")
	for _, tx := range ledger.Transactions(){
		fmt.Printf("  %+v\n", tx)
	}

	fmt.Println("\nПроверка, что превышающая транзакция ID=2 не сохранена")
	for _, tx := range ledger.Transactions(){
		if tx.ID == "2" {
			fmt.Println("ОШИБКА: транзакция ID=2 всё же сохранена")
			return
		}
	}
	fmt.Println("YES! -> Транзакции ID=2 нет в списке")

	fmt.Println("\nЗагрузка бюджетов из budgets.json...")
	ensureSampleBudgetsFile()

	file, err := os.Open("budgets.json")
	if err != nil{
		log.Printf("Не удалось открыть budgets.json: %v", err)
		return
	}
	defer file.Close()

	if err := ledger.LoadBudgets(bufio.NewReader(file)); err != nil{
		log.Printf("Ошибка загрузки бюджетов: %v", err)
		return
	}
	fmt.Println("Бюджеты успешно загружены")
}