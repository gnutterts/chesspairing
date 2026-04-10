---
title: "Formaatspecificaties"
linkTitle: "Formaten"
weight: 80
description: "Specificaties van bestandsformaten — TRF16, TRF-2026-extensies, JSON-schema's en configuratie."
---

Deze sectie documenteert de bestandsformaten die chesspairing gebruikt voor gegevensuitwisseling van toernooien, configuratie van engines en gestructureerde uitvoer.

Chesspairing gebruikt TRF16 als primair bestandsformaat voor gegevensuitwisseling van toernooien. TRF16 is een tekstformaat met vaste breedte, gedefinieerd door de FIDE voor toernooiverslagbestanden. Het `trf`-pakket biedt een volledige lezer en schrijver met bidirectionele conversie van en naar de interne `TournamentState`-representatie.

TRF-2026 breidt TRF16 uit met extra recordtypes en systeemspecifieke velden. Deze extensies voegen headercodes toe voor het totale aantal ronden, beginkleur, scoresystemen en tiebreaker-definities, evenals datarecords voor afwezigheden, acceleratie, verboden paren en teamgegevens. Het `trf`-pakket verwerkt zowel TRF16- als TRF-2026-documenten transparant.

De CLI produceert gestructureerde JSON-uitvoer voor indelingen, ranglijsten, validatie, versie-informatie en tiebreaker-overzichten. Alle JSON-uitvoer gebruikt inspringen met 2 spaties.

Engine-opties worden geconfigureerd via een `map[string]any` sleutel-waardeformaat dat afkomstig kan zijn uit TRF-velden, JSON-configuratie of CLI-flags. Het `generate`-subcommando accepteert ook een RTG-configuratiebestand met `key=value`-syntaxis.

| Pagina                                | Beschrijving                                                          |
| ------------------------------------- | --------------------------------------------------------------------- |
| [TRF16](trf16/)                       | Het FIDE TRF16-formaat -- regeltypes, spelersrecords, ronderesultaten |
| [TRF-2026-extensies](trf-extensions/) | Systeemspecifieke XX-velden voor configuratie van indelingsengines      |
| [JSON-schema's](json-schemas/)        | JSON-uitvoerschema's voor CLI-commando's                              |
| [Configuratie](configuration/)        | Sleutel-waardeconfiguratie voor engine-fabrieken en RTG               |
