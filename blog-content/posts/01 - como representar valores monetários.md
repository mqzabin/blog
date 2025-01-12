
+++
title = 'Como representar valores monetários'
date = 2024-07-30T19:30:59-03:00
draft = true
math = true
+++

# Introdução

Trabalhei alguns anos no desenvolvimento de softwares que gerenciavam dinheiro, e especifiquei uma série de sistemas que precisavam armazenar e manipular valores monetários.

Em toda especificação, sempre surgia a necessidade de definir como o dinheiro seria representado: Em código, no banco de dados e em nossas APIs. Por ser tratar de um problema básico, muitas vezes essa decisão era abordada com ingenuidade.

> "Como vamos armazenar os valores dos produtos?"

Usar *float* (pontos flutuantes), inteiros como centavos ou alguma estrutura pronta de alguma biblioteca? Como será representado no banco de dados? E no JSON de entrada/saída da API?

Pensando em valores do dia-a-dia e fazendo alguns testes, armazenar $ 2.00 parece funcionar para quase qualquer tipo de estrutura numérica.

Podemos simplesmente ignorar que algum dia alguém queira perguntar em um relatório "quanto dinheiro foi movimentado em nosso projeto?"; ou que o problema que queremos resolver mude de escala numérica drasticamente; ou que vamos ter que fazer milhares de operações aritméticas em uma chamada do usuário.

Ninguém consegue prever o futuro e não há solução definitiva. Porém é importante conhecer as opções disponíveis, e quais os pontos positivos e negativos de cada uma.

Esse artigo busca abordar quais são as principais opções, e quais suas contrapartidas. Para isto, vamos dividir o problema em três diferentes camadas:

- Em memória RAM (i.e. no código).
- Nos contratos das APIs.
- No banco de dados.

# Em memória

Quando representamos valores numéricos em memória, temos algumas preocupações:

- Esse dado conseguirá representar o intervalo de valores que o sistema precisa representar?
- Quão a performance das operações aritméticas nesse tipo de dado?

## Pontos flutuantes

Uma resposta muito comum (e ingênua) que recebo é o uso de pontos flutuantes. As linguagens de programação geralmente tem um tipo primitivo para representação de pontos flutuantes, por exemplo:

- Golang: `float32` e `float64`.
- JavaScript: `number`.
- C/Java: `float`  e `double`. 

Pontos flutuantes são leves, as operações são rápidas, e tudo fica na *stack*. Pode parecer uma boa ideia! /s

Para entender a escolha, precisamos explorar o que significa "pontos flutuante". A estrutura de pontos flutuantes utilizam uma representação parecida com a "notação científica", que aprendemos no ensino médio. Isto é, o valor é guardado em dois componentes:

- **(M)antissa:** Valor significativo do numéro.
- **(E)xpoente:** Aonde o ponto será colocado na mantissa.

$$ M \times 10^{E} $$

A nomenclatura de "ponto flutuante" surge do fato de que o valor de $E$, que define aonde o ponto decimal será colocado, pode "flutuar" para mais ou menos.

A priori, essa representação não oferece nenhum problema, já que a notação científica consegue representar qualquer número com uma precisão arbitrária. 

O problema surge quando trazemos essa representação para dentro de um número limitado de memória. Por exemplo, quando lidamos com os 64 bits disponíveis em um `float64` (ou `double`), a disposição de memória é a seguinte:

- 1 bit para o sinal (positivo/negativo).
- 11 bits para o expoente.
- 52 bits para a mantissa/dígitos significativos.

Totalizando os 64 bits disponíveis.

Os 11 bits de expoente trazem um intervalo de cobertura de valores imenso, podendo representar valores com 308 dígitos decimais (antes ou depois da vírgula).

O problema existe nos 52 bits disponíveis para armazenamento da mantissa, que podem armazenar aproximadamente ~16 dígitos decimais. Isso significa que a partir dessa quantidade de digitos, nenhum valor poderá ser garantidamente representado. 16 dígitos pode parecer grande o suficiente, mas lembre-se que 2 desses dígitos já são designados 

Veja o exemplo abaixo, em *Golang*, onde tentamos atribuir 30 digitos significativos à um `float64`.

```go
package main

import (
	"fmt"
)

func main() {
	var n float64

	// 20 dígitos antes e 10 dígitos depois da vírgula.
	n = 12345678901234567890.1234567890

	fmt.Printf("%.20f\n", n)
	// 12345678901234567168.00000000000000000000
}
```

Conseguimos ver que os digitos menos significativos acabam sendo zerados, ou tendo seus valores alterados dado as limitações de memória.

Outro fator importante, é que estamos tentando utilizar valores decimais em cima de um ponto flutuante binário. Assim como não conseguimos representar a fração 10/3 em valores decimais, também temos restrições na representação binária.

Por se tratar da base 2 em vez da base 10, temos ainda mais dificuldade na representação de valores "quebrados", aparecendo erros grotescos aonde menos esperamos, com impactos que podem estar na casa das centenas de milhares. Por exemplo:

```go
package main

import (
	"fmt"
)

func main() {
	var a float64

	a = 12_345_000_000_000_000_000_000.0

	fmt.Printf("%.20f\n", a)
	// 12344999999999999737856.00000000000000000000
}
```

E perceba que neste caso temos apenas 5 digitos significativos!!



## Números inteiros

## Decimais de precisão arbitrária

# No contratos das APIs

# No banco de dados

