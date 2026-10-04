---
title: "Teamscoring"
weight: 35
---

Teamtoernooien gebruiken twee scores: matchpunten (MP) en partijpunten (GP).
Gebruik `team` als `system` in `ScoringConfig`. MP is standaard de primaire
score; met `primaryScore` op `game` wordt GP de scalaire `PlayerScore.Score`.
Beide waarden staan ook in `PlayerScore.Team`.

`RoundData.Matches` bevat een teammatch. De bordresultaten bepalen GP; als
borden ontbreken bevat `MatchData.Result` de opgegeven totalen. De standaard
matchschaal is 2 punten voor winst, 1 voor remise en 0 voor verlies.
`pointMatchWin`, `pointMatchDraw` en `pointMatchLoss` configureren die schaal.
De bordpunten gebruiken de standaard sleutels `pointWin`, `pointDraw` en
`pointLoss`. Team-byes horen in `RoundData.TeamByes` (bij voorkeur boven `Byes`); hun
partijpunten gebruiken de remise- of winstwaarde op elk bord, met het aantal
borden van de grootste match in het evenement.

TRF-2026-records `013` koppelen spelers aan teams; de volgorde bepaalt het
teamnummer, want record `013` heeft geen teamnummerveld. Gedetailleerde
`801`-records worden matches met borden en eenvoudige `802`-records worden
matches met expliciete GP-totalen. Record `320` bevat team-byes die door de
indeling zijn toegekend en wordt omgezet naar `RoundData.TeamByes`. Record
`162` configureert bordpunten. Teamnamen en lidmaatschap dat alleen in `310`
staat, worden niet in `TournamentState` bewaard; gebruik `013`-lidmaatschap
bij conversie van team-TRF-bestanden. Een team zonder match of bye krijgt geen
impliciete afwezigheidsstraf, omdat C.04.6 er geen definieert.

## Teamtiebreaks

Het team-specifieke register biedt de C.07-varianten `mpvgp`, `emmsb`,
`emmsb-cut1`, `emgsb` en `egmsb`, plus `eggsb`, `buchholz-mp`,
`buchholz-mp-cut1`, `board-count`, `top-board-results` en
`bottom-board-elimination`. Zij zijn beschikbaar via `tiebreaker.Get` en
staan in het CLI-commando `tiebreakers`.
