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

TRF-2026-record `362` configureert `pointMatchWin`, `pointMatchDraw` en `pointMatchLoss`; de standaard blijft 2-1-0. Record `352` configureert zowel het aantal borden als de kleurvolgorde van het thuisspelende team per bord (`WWBB`, bijvoorbeeld). Record `310` koppelt spelers aan teams met expliciete nummers; `013` is de oudere terugvaloptie en wordt nog geschreven. Alleen bij `013` bepaalt de recordvolgorde het teamnummer. Gedetailleerde
`801`-records worden matches met borden en eenvoudige `802`-records worden
matches met expliciete GP-totalen. Record `320` bevat team-byes die door de
indeling zijn toegekend en wordt omgezet naar `RoundData.TeamByes`. Record
`162` configureert bordpunten. Teamnamen en spelers uit `310` blijven beschikbaar in `TournamentState`. `300` wijzigt de bordvolgorde en `330` maakt ontbrekende forfaitwedstrijden aan. `299`-regels W/D/L worden toegepast (en moeten met `362` overeenkomen); F/H/Z-, `+`/`-`- en regels zonder type worden alleen bewaard, omdat de teamscorer geen afzonderlijke bye-waarden heeft. Een team zonder match of bye krijgt geen
impliciete afwezigheidsstraf, omdat C.04.6 er geen definieert.

## Teamtiebreaks

Het team-specifieke register biedt de C.07-varianten `mpvgp`, `emmsb`,
`emmsb-cut1`, `emgsb` en `egmsb`, plus `eggsb`, `buchholz-mp`,
`buchholz-mp-cut1`, `board-count`, `top-board-results` en
`bottom-board-elimination`. Zij zijn beschikbaar via `tiebreaker.Get` en
staan in het CLI-commando `tiebreakers`.
