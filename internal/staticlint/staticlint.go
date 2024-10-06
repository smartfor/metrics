// Package staticlint аггрегирует множество анализаторов статического кода в один multichecker.
// Список анализаторов
//
// # Описание всех анализаторов:
// 1. asmdecl - анализирует объявления ассемблерных функций
// 2. assign - анализирует присваивания
// 3. atomic - анализирует использование пакета sync/atomic
// 4. atomicalign - анализирует выравнивание переменных для атомарных операций
// 5. bools - анализирует использование булевых переменных
// 6. buildssa - строит SSA форму программы
// 7. buildtag - анализирует использование build тегов
// 8. cgocall - анализирует вызовы C функций
// 9. composite - анализирует использование составных литералов
// 10. copylock - анализирует использование копирования мьютексов
// 11. ctrlflow - анализирует поток управления
// 12. deepequalerrors - анализирует использование reflect.DeepEqual для ошибок
// 13. directive - анализирует использование директив
// 14. errorsas - анализирует использование ошибок в качестве значений
// 15. fieldalignment - анализирует выравнивание полей структур
// 16. findcall - анализирует вызовы функций
// 17. framepointer - анализирует использование указателей на кадры стека
// 18. httpresponse - анализирует использование http.ResponseWriter
// 19. ifaceassert - анализирует приведение интерфейсов
// 20. inspect - анализирует программу
// 21. loopclosure - анализирует замыкания в циклах
// 22. lostcancel - анализирует утечки контекста
// 23. nilfunc - анализирует вызовы nil функций
// 24. nilness - анализирует использование nil
// 25. pkgfact - анализирует факты о пакетах
// 26. printf - анализирует использование функции fmt.Printf
// 27. reflectvaluecompare - анализирует сравнение значений reflect.Value
// 28. shadow - анализирует затенение переменных
// 29. shift - анализирует сдвиги
// 30. sigchanyzer - анализирует использование каналов и сигналов
// 31. sortslice - анализирует сортировку срезов
// 32. stdmethods - анализирует использование методов стандартной библиотеки
// 33. stringintconv - анализирует преобразование строк в целые числа
// 34. structtag - анализирует использование тегов структур
// 35. testinggoroutine - анализирует использование горутин в тестах
// 36. tests - анализирует тесты
// 37. timeformat - анализирует форматирование времени
// 38. unmarshal - анализирует использование функции Unmarshal
// 39. unreachable - анализирует недостижимый код
// 40. unsafeptr - анализирует использование unsafe.Pointer
// 41. unusedresult - анализирует неиспользуемые результаты
// 42. unusedwrite - анализирует неиспользуемые записи
// 43. usesgenerics - анализирует использование обобщений
// 44. ineffassign - анализирует неэффективные присваивания
// 45. go-critic - анализирует код на предмет соответствия рекомендациям. Использует анализаторы библиотеки go-critic
// 46. noexit - проверяет наличие вызовов os.Exit в пакете main в методе main
package staticlint

import (
	"strings"

	gocricit "github.com/go-critic/go-critic/checkers/analyzer"
	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func Run() {
	checks := make([]*analysis.Analyzer, 0)
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			checks = append(checks, v.Analyzer)
		}
	}

	for _, v := range stylecheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	checks = append(checks,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		atomicalign.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
		ineffassign.Analyzer,
		gocricit.Analyzer,
		NoExitAnalyzer,
	)

	multichecker.Main(checks...)
}
