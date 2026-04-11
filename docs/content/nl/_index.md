---
title: ChessPairing
description: "Go-bibliotheek en CLI voor het indelen, scoren en rangschikken van schaaktoernooien volgens FIDE-reglementen."
---

{{< blocks/cover title="ChessPairing" image_anchor="top" height="med" color="dark" >}}

<div class="mx-auto">
  <a class="btn btn-lg btn-primary me-3 mb-4" href="/docs/">
    Documentatie
  </a>
  <a class="btn btn-lg btn-secondary me-3 mb-4" href="https://github.com/gnutterts/chesspairing">
    GitHub
  </a>
  <p class="lead mt-4">Een Go-module voor het indelen, scoren en rangschikken van schaaktoernooien.</p>
</div>
{{< /blocks/cover >}}

{{% blocks/lead color="primary" %}}

Go-bibliotheek en CLI-tool voor het indelen van schaaktoernooien volgens FIDE-reglementen. Ondersteunt Zwitsers (Nederlands, Burstein, Dubov, Lim, Dubbel-Zwitsers, Team), Round-Robin en Keizer. Geen externe afhankelijkheden.

{{% /blocks/lead %}}

{{< blocks/section color="white" type="row" >}}

{{% blocks/feature icon="fa-chess" title="Spelers en Arbiters" url="/docs/getting-started/for-arbiters/" %}}
Hoe indelingen werken, wat de tiebreakers meten, en hoe je de uitvoer leest.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-terminal" title="CLI" url="/docs/getting-started/cli-quickstart/" %}}
Installeer de tool, voer een TRF-bestand in, krijg indelingen terug. Werkt als drop-in voor bbpPairings en JaVaFo.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-code" title="Go API" url="/docs/getting-started/go-quickstart/" %}}
`go get github.com/gnutterts/chesspairing` en deel programmatisch in. Geen afhankelijkheden, veilig voor gelijktijdig gebruik.
{{% /blocks/feature %}}

{{< /blocks/section >}}
