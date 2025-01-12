+++
title = 'Testes Fuzzy em Golang - Além da página 1'
date = 2025-01-12T12:00:00-03:00
draft = false
math = true
+++

# Introdução

O Go 1.18 introduziu o suporte à [testes fuzzy](https://go.dev/doc/security/fuzz/), e como sempre, uma série de blogposts e vídeos foram produzidos explorando a nova funcionalidade.

O propósito dos testes fuzzy é utilizar as teorias matemáticas/computacionais da [lógica fuzzy](https://pt.wikipedia.org/wiki/L%C3%B3gica_difusa) para explorar caminhos de execução do seu código automaticamente, sem necessariamente precisar de algum input escrito manualmente.

Mas se você já leu alguma dessas explorações de testes fuzzy em Go, já conhece qual o propósito dos testes fuzzy e provavelmente notou que a maioria dos exemplos giram em torno de testar funções que possuem alguma propriedade matemática/computacional especial.

Por exemplo, testar inversão de strings, onde aplicar a função duas vezes deve retornar o valor original (i.e. $f(f(x)) = x$ ou $f^{-1}(x) = f(x)$):

```go
package main

import (
    "testing"
    "unicode/utf8"
)

func Reverse(s string) (string, error) {
    if !utf8.ValidString(s) {
        return s, errors.New("input is not valid UTF-8")
    }
    r := []rune(s)
    for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
        r[i], r[j] = r[j], r[i]
    }
    return string(r), nil
}

func FuzzReverse(f *testing.F) {
    f.Fuzz(func(t *testing.T, orig string) {
        rev, err1 := Reverse(orig)
        if err1 != nil {
            return
        }
        doubleRev, err2 := Reverse(rev)
        if err2 != nil {
            return
        }
        if orig != doubleRev {
            t.Errorf("Before: %q, after: %q", orig, doubleRev)
        }
        if utf8.ValidString(orig) && !utf8.ValidString(rev) {
            t.Errorf("Reverse produced invalid UTF-8 string %q", rev)
        }
    })
}
```

Ou então, testar a comutatividade de uma operação (i.e. $f(a,b) = f(b, a)$), como na soma ou multiplicação de dois números:

```go
package main  
  
import "testing"  
  
func FuzzCommutativeOperations(f *testing.F) {  
    f.Fuzz(func(t *testing.T, a, b uint64) {  
       if a+b != b+a {  
          t.Errorf("a+b: %d, b+a: %d", a+b, b+a)  
       }  
  
       if a*b != b*a {  
          t.Errorf("a*b: %d, b*a: %d", a*b, b*a)  
       }  
    })  
}
```

(o código acima é só um exemplo sem utilizar bibliotecas externas, mas essas operações poderiam vir de um pacote, e operar alguma estrutura especial)

O problema é que essas propriedades são muito raras em problemas reais que tentamos resolver com desenvolvimento de software, fazendo com que testes fuzzy pareçam meio inúteis.

Entretanto, a ideia deste artigo é explorar algumas formas diferentes de utilizar testes fuzzy. Aumentar o repertório de usos faz com que seja mais fácil enxergar onde os softwares "mais clássicos" podem se encaixar com API fuzzy do Go.

# Superando as limitações da API do Go

O método `(*testing.F).Fuzz` recebe um parâmetro do tipo `interface{}`/`any`. Esse parâmetro deve ser uma função que recebe um `*testing.T` e uma sequência customizável de parâmetros de tipos primitivos.

Os tipos aceitos como parâmetros são:
- `string`, `[]byte`
- `int`, `int8`, `int16`, `int32`/`rune`, `int64`
- `uint`, `uint8`/`byte`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `bool`

A primeira vista, parece muito dificil encaixar `structs` customizadas de um software comum dentro desses tipos primitivos.

Porém, podemos compor esses tipos, transformando os parametros em valores significativos para nosso software.

Por exemplo, digamos que a aplicação possua um enumerador de `PaymentMethod` como:

```go
type PaymentMethod uint8

const (
	PaymentMethodCreditCard PaymentMethod = iota
	PaymentMethodDebitCard
	PaymentMethodCash
)

var (
	ErrInvalidPaymentMethod = errors.New("invalid payment method")

	StringToPaymentMethod = map[string]PaymentMethod{
		 "credit_card": PaymentMethodCreditCard,
		 "debit_card": PaymentMethodDebitCard,
		 "cash": PaymentMethodCash,
	}
	PaymentMethodToString = map[PaymentMethod]string{
		PaymentMethodCreditCard: "credit_card",
		PaymentMethodDebitCard: "debit_card",
		PaymentMethodCash: "cash",
	}
)

func NewPaymentMethodFromString(s string) (PaymentMethod, error) {
	pm, valid := StringToPaymentMethod[s]
	if !valid {
		return 0, ErrInvalidType
	}

	return pm, nil
}

func (pm PaymentMethod) String() string {
	return PaymentMethodToString[pm]
}

```

Se você deseja gerar um `PaymentMethod` como input do teste fuzzy, você pode fazer algo como:

```go
func FuzzFromEnum(f *testing.F) {
	paymentMethods := slices.SortedStableFunc(maps.Keys(PaymentMethodToString), func(a PaymentMethod, b PaymentMethod) int {
		return int(a) - int(b)
	})


	getPaymentMethod := func(t *testing.T, seed uint) PaymentMethod {
		t.Helper()

		return PaymentMethod(seed % len(paymentMethods))
	}


	f.Fuzz(func(t *testing.T, enum uint) {
		pm := getPaymentMethod(t, enum)

		// Do some stuff with pm
	})
}

```

*(os detalhes de implementação do código não importam para a mensagem, existem infinitas formas de abordar o problema)*

Obviamente, isso pode ser feito com slices, mapas, switches, etc.

Ou caso necessite de uma string numérica, o teste pode:

```go
func FuzzyNumericString(f *testing.F) {  
    f.Fuzz(func(t *testing.T, a, b uint64) {  
       numericString := strconv.FormatUint(a, 10)  
       numericString += strconv.FormatUint(b, 10)  
  
       // Do some stuff with numericString.  
    })  
}
```

E claro, é possível aplicar as mais variadas instruções para construir alguma string dentro de um formato específico.

Entretanto, como essas transformações criam novos caminhos de execução (que o teste fuzzy está tentando explorar), possivelmente estamos perdendo algum grau de eficiência para atingir os formatos esperados pelo nosso software. But still...

# O pacote [`mqzabin/fuzzdecimal`](https://github.com/mqzabin/fuzzdecimal)

No mundo financeiro é muito comum que representemos números decimais utilizando estruturas complexas que garantam uma precisão absoluta nas operações matemáticas.

Recentemente me aventurei em criar um pacote para representar e realizar operações com números decimais de precisão arbitrária, utilizando zero alocações (esse é o selling point do pacote).

Porém, quando estamos lidando com valores numéricos, existe um grande desafio na criação de testes, dado que existem infinitas possibilidades de input. Mesmo sendo possível analisar os caminhos de execução e escrever inputs que cubram esses caminhos, quando chegamos em operações não-básicas que utilizam métodos iterativos (por exemplo divisão e exponenciação), é muita loucura manter o tracking de todos os possíveis caminhos.

Essa dificuldade (de forma isolada) parece ser facilmente resolvida pelos testes fuzzy do Go. Porém, mesmo que existam inputs auto-gerados que percorram caminhos diferentes do código, como será feita a detecção de erros de cálculo?

É necessário ter um valor esperado da operação, e a API fuzzy não parece ajudar nesse aspecto.

Porém, como estamos lidando com matemática e as operações são bem definidas em qualquer lugar do mundo, posso utilizar outro pacote de decimais de precisão arbitrária como valor esperado da minha operação.

Na companhia onde trabalho, é bem comum utilizarmos o `github.com/shopspring/decimal`. Esse pacote é bem conhecido, usado e testado, e possuía todas as operações que estavam implementadas no pacote que eu estava desenvolvendo.

Foi ai então que decidi criar testes fuzzy comparando resultados da minha biblioteca com a `shopspring/decimal`. A lógica dos testes é a seguinte:

- Os testes recebem parâmetros `uint64` suficientes para criar X números suficientemente grandes.
- Os `uint64` são combinados em strings numéricas que podem ser convertidas para decimais de precisão arbitrária de ambas as bibliotecas (pelo meu pacote, e pelo `shopspring/decimal`).
- Efetuo a operação testada pelo meu pacote e obtenho um resultado A.
- Efetuo a operação testada pelo `shopspring/decimal` e obtenho um resultado B.
- Normalizo ambos os resultados para uma string de precisão fixa (dado que minha biblioteca não tem precisão dinâmica/"infinita").
- Comparo os resultados e falho o teste fuzzy caso eles tenham divergências.

Com esses passos, consegui testar todas as operações das funções expostas pelo meu pacote sem escrever sequer um caso de teste.

Depois de evoluir meu pacote de decimais e iniciar outros projetos (relacionados à matemática), notei que precisaria extrair essa lógica de testes fuzzy com decimais de precisão arbitrária para ser reutilizado nos outros projetos.

Foi então que escrevi o pacote [github.com/mqzabin/fuzzdecimal](https://github.com/mqzabin/fuzzdecimal), que expõe uma API simples para realizar testes com decimais de precisão arbitrária sem muito setup e funções auxiliares.

Exemplo de como ficou os testes de adição e subtração do meu pacote:

```go
  
func FuzzAddSub(f *testing.F) {  
    parseDecimal := func(t *fuzzdecimal.T, s string) (Currency, error) {  
       t.Helper()  
  
       return NewFromString(s)  
    }  
  
    parseShopspringDecimal := func(t *fuzzdecimal.T, s string) (decimal.Decimal, error) {  
       t.Helper()  
  
       return decimal.NewFromString(s)  
    }  
  
    fuzzdecimal.Fuzz(f, 2, func(t *fuzzdecimal.T) {  
       fuzzdecimal.AsDecimalComparison2(t, "Add", parseDecimal, parseShopspringDecimal,  
          func(t *fuzzdecimal.T, x1, x2 decimal.Decimal) (string, error) {  
             t.Helper()  
  
             return x1.Add(x2).Truncate(currencyDecimalDigits).String(), nil  
          },  
          func(t *fuzzdecimal.T, x1, x2 Currency) string {  
             return x1.Add(x2).String()  
          },  
       )  
  
       fuzzdecimal.AsDecimalComparison2(t, "Sub", parseDecimal, parseShopspringDecimal,  
          func(t *fuzzdecimal.T, x1, x2 decimal.Decimal) (string, error) {  
             t.Helper()  
  
             return x1.Sub(x2).Truncate(currencyDecimalDigits).String(), nil  
          },  
          func(t *fuzzdecimal.T, x1 Currency, x2 Currency) string {  
             return x1.Sub(x2).String()  
          },  
       )  
    }, fuzzdecimal.WithAllDecimals(  
       fuzzdecimal.WithSigned(),  
       // The a+b and a-b operations will at most add 1 digit to the greatest number between a and b.  
       // So, we should ensure that the greatest number has at most naturalMaxLen-1 digits.
       fuzzdecimal.WithMaxSignificantDigits(naturalMaxLen-1),  
       fuzzdecimal.WithDecimalPointAt(currencyDecimalDigits),  
    ))  
}

```

Apesar da implementação utilizar algumas reflections, como o escopo é bem pequeno (somente números decimais), o código ficou fácil de entender, e a API ficou bem limpa dado que a lógica de reflections ficou completamente abstraída.

Leia o README do projeto para conhecer as demais funções.

# Possibilidades futuras (sonhando alto)

Seguindo a mesma lógica do pacote [github.com/mqzabin/fuzzdecimal](https://github.com/mqzabin/fuzzdecimal), é possível desenvolver construtores de estruturas complexas mais genéricas.

É possível criar um pacote que tenha uma API que suporte transformações comuns, por exemplo:

- UUIDs
- Strings alfanuméricas
- Números decimais
- Enums
- Números primitivos com intervalos definidos.

Isso seria muito útil para diversos testes nas camadas de domínio de nossas aplicações.

Claro que tudo isso é apenas um esboço, mas podemos pensar em uma API que seja relativamente simples, e que forneça todas essas facilidades. Considerando o novo pacote como `newfuzzlib`, poderíamos ter algo como:

```go
func FuzzUserPurchase(f *testing.F) {
	// Defined at your domain packages.
	type (
		User struct {
			ID 						uuid.UUID
			Name 					string
			Age  					int
			TotalPurchasesAmount 	shopspring.Decimal
		}
		Purchase struct {
			ID 		uuid.UUID
			UserID 	uuid.UUID
			Amount 	shopspring.Decimal
		}
	)

	newfuzzlib.Fuzz(f,
		func(t *testing.T, user User, purchase Purchase) {
			// Do some stuff with user and purchase.
		},
		// User
		newfuzzlib.ParamStruct(1,
			newfuzzlib.FieldUUID("ID",
				newfuzzlib.UUID(),
			),
			newfuzzlib.FieldString("Name",
				newfuzzlib.MaxLength(100),
				newfuzzlib.AlphaNumeric(),
			),
			newfuzzlib.FieldInt("Age",
				newfuzzlib.Min(0),
				newfuzzlib.Max(100),
			),
			newfuzzlib.FieldDecimal("TotalPurchasesAmount",
				func(t *testing.T, str string) (shopspring.Decimal, error) {
					return shopspring.NewFromString(str)
				},
				newfuzzlib.MaxSignificantDigits(10),
				newfuzzlib.MaxDecimalDigits(2),
			),
		),
		// Purchase
		newfuzzlib.ParamStruct(2,
			newfuzzlib.FieldUUID("ID",
				newfuzzlib.V7(),
			),
			newfuzzlib.FieldUUID("UserID",
				newfuzzlib.SameAsStructField(1, "ID"),
			),
			newfuzzlib.FieldDecimal("Amount",
				func(t *testing.T, str string) (shopspring.Decimal, error) {
					return shopspring.NewFromString(str)
				},
				newfuzzlib.MaxSignificantDigits(10),
				newfuzzlib.MaxDecimalDigits(2),
			),
		),
	)
}

```

Onde temos as estruturas `User` e `Purchase` sendo parâmetros do teste fuzzy, tendo as restrições dos seus campos sendo explicitamente definidas (através do índice do parâmetro).

Essas ideias ficam para o futuro. Caso algum dia tenha tempo para esse novo projeto, provavelmente farei algo na linha da API que exemplifiquei acima.