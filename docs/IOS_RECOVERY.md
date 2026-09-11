# Begrenzte iOS-Wiederherstellung nach Netzunterbrechungen

Status: **auf `testing` implementiert und statisch geprüft; Docker-Builds und die iPhone-Laufzeitprüfung auf dem Zielserver stehen noch aus**.

Dieses Runbook ist für einen Linux-Zielserver gedacht, auf dem nur Git, Bash, Docker Engine und `docker compose` vorhanden sein müssen. Node.js, npm, Go und Python werden auf dem Host **nicht** benötigt. Die automatisierbaren Prüfungen laufen in kurzlebigen Containern über [`docker-compose.validation.yaml`](../docker-compose.validation.yaml).

Der Checkpoint muss gemeinsam mit der Drei-Rollen-Prüfung aus [`VIEW_ONLY_SHARING.md`](VIEW_ONLY_SHARING.md) auf demselben Commit und demselben Neko-Image erfolgen. Ein bestandenes Runbook ist kein Nachweis für das andere.

## Was geprüft wird

Die Implementierung trennt drei Zustände:

1. **Bestehender Peer:** ICE darf sich acht Sekunden lang selbst erholen. Bei Erfolg bleiben WebSocket, Peer und Server-Session bestehen.
2. **Neue Anwendungssession:** Ist der alte Peer nicht mehr verwendbar, werden Socket, Peer, Data Channel, Timer, Kandidaten und Medienreferenzen entfernt. Danach folgen höchstens vier serielle Loginversuche nach 1, 2, 5 und 10 Sekunden.
3. **Safari-Play-Fallback:** Ist das Netz bereits wiederhergestellt und Media vorhanden, darf Safari noch eine Play-Geste verlangen. Dieser Tap ist eine Autoplay-Vorgabe und kein fehlgeschlagener Reconnect.

Initiale Loginfehler, explizites Logout, Demo-Modus und ein serverseitiges `system/disconnect` dürfen keine automatische Loginserie auslösen. Veraltete Callbacks dürfen eine neue Verbindung nicht verändern. Die Media-Element-Wiederherstellung bleibt begrenzt und verwendet niemals `video.load()`.

## Übersicht des Ablaufs

| Block | Ausführung | Ergebnis |
| --- | --- | --- |
| 0 | Git/Bash auf dem Server | exakter Commit und Ergebnisordner |
| 1 | Docker Compose | Client: `npm ci`, Recovery-Test, TypeScript-Lint, Vite-Build |
| 2 | Docker Compose | fokussierte Go-Tests und Server-Build |
| 3 | vorhandenes `./build` | lokaler Base-/Brave-Image-Build; intern vollständig containerisiert |
| 4 | Docker Compose | adaptive Zielbereitstellung und Startnachweis |
| A–C | reales iPhone plus zwei gesunde Zuschauer | Same-Peer-, Replacement- und Exhaustion-Test |

Die Blöcke 0–4 werden für den gemeinsamen iOS/View-only-Checkpoint nur einmal ausgeführt.

## Sicherheits- und Evidenzregeln

- Niemals `.env`, Passwörter, den View-only-Token oder die Ausgabe eines nicht stillen `docker compose config` weitergeben.
- Ausschließlich `docker compose ... config --quiet` verwenden.
- Ergebnisdateien außerhalb des Repositorys ablegen; Session-IDs und Hostdetails können sensibel sein.
- Alle Befehle aus dem Repository-Stamm ausführen.
- Bei einem Fehler nicht weiterlaufen oder Werte verändern, sondern den vollständigen Block-Output übergeben.
- Die folgenden Befehle verwenden absichtlich das bereits akzeptierte adaptive Overlay, damit die Regression der gesunden Zuschauer mitgeprüft wird. Wird bewusst nur das Basisprofil verwendet, müssen bei **allen** Compose-Befehlen `-f docker-compose.adaptive.yaml` und die tierbezogenen Akzeptanzpunkte entfallen; dies ist im Ergebnis zu vermerken.

## Block 0 — exakten Checkpoint vorbereiten

Diesen Block vollständig kopieren. Die Shell für die folgenden Blöcke geöffnet lassen, damit die exportierten Variablen erhalten bleiben.

```bash
set -o pipefail
git switch testing
git pull --ff-only origin testing
git status --short --branch

export NEKO_VALIDATION_COMMIT="$(git rev-parse HEAD)"
export NEKO_RESULT_DIR="../neko-checkpoint-${NEKO_VALIDATION_COMMIT:0:12}-$(date -u +%Y%m%dT%H%M%SZ)"
export NEKO_ROLLBACK_IMAGE="my-neko/brave:pre-ios-view-only-${NEKO_VALIDATION_COMMIT:0:12}"
mkdir -p "$NEKO_RESULT_DIR"

{
  date -u +'%Y-%m-%dT%H:%M:%SZ'
  git status --short --branch
  git rev-parse HEAD
  git rev-parse origin/testing
  echo "rollback_image=$NEKO_ROLLBACK_IMAGE"
  docker --version
  docker compose version
} 2>&1 | tee "$NEKO_RESULT_DIR/00-provenance.txt"
```

Erwartung:

- Branch `testing`;
- Arbeitsbaum sauber;
- `HEAD` und `origin/testing` identisch.

Wenn `git status --short --branch` zusätzliche Dateiänderungen zeigt oder die beiden Hashes abweichen: stoppen und `00-provenance.txt` übergeben.

## Block 1 — Client vollständig im Container prüfen

Der Quellbaum wird read-only eingebunden und in ein temporäres Container-Dateisystem kopiert. `node_modules` und `dist` landen nicht auf dem Host.

```bash
(
  set -e
  docker compose -f docker-compose.validation.yaml config --quiet
  docker compose -f docker-compose.validation.yaml pull client-checks
  docker compose -f docker-compose.validation.yaml run --rm client-checks
) 2>&1 | tee "$NEKO_RESULT_DIR/01-client-checks.txt"
```

Erwartung: `npm ci`, `npm test`, `npm run lint` und `npm run build` enden ohne Fehler. Dieser eine Block deckt alle Clientbefehle ab; nichts davon muss auf dem Host installiert sein.

## Block 2 — Servertests und Server-Build im Container

Der Compose-Service baut das Repository-`server/Dockerfile`, führt danach die fokussierten Pakete aus und startet `./build` nochmals im kurzlebigen Prüfcontainer.

```bash
(
  set -e
  docker compose -f docker-compose.validation.yaml build --pull server-checks
  docker compose -f docker-compose.validation.yaml run --rm server-checks
) 2>&1 | tee "$NEKO_RESULT_DIR/02-server-checks.txt"
```

Geprüfte Pakete:

```text
./pkg/types
./pkg/auth
./internal/member/multiuser
./internal/session
./internal/http/legacy
./internal/websocket
./internal/webrtc
```

Erwartung: sämtliche `go test`-Pakete melden Erfolg und der anschließende Server-/Plugin-Build endet ohne Fehler. Go muss auf dem Host nicht vorhanden sein.

## Block 3 — Rollback-Image sichern und Zielimages bauen

Die folgenden Befehle gehen vom dokumentierten Image `my-neko/brave:latest` aus und sichern es unter einem commitbezogenen Rollback-Tag. Falls `.env` absichtlich einen anderen `NEKO_IMAGE`-Namen verwendet, muss der Quellname ersetzt und `NEKO_ROLLBACK_IMAGE` vor diesem Block passend neu gesetzt werden.

`./build` ist der vorhandene Repository-Wrapper. Er startet die erforderlichen Go-/Node-/Docker-Buildschritte selbst in Containern; npm oder Go werden dadurch nicht auf dem Host installiert.

```bash
(
  set -e
  docker image inspect --format 'source_image={{.RepoTags}} id={{.Id}} created={{.Created}}' my-neko/brave:latest
  docker image tag my-neko/brave:latest "$NEKO_ROLLBACK_IMAGE"
  ./build my-neko/base:latest -y
  ./build my-neko/brave:latest -y
  docker image inspect --format 'image={{.RepoTags}} id={{.Id}} created={{.Created}}' my-neko/brave:latest
) 2>&1 | tee "$NEKO_RESULT_DIR/03-image-build.txt"
```

Erwartung: beide Builds erfolgreich; der abschließende Inspect zeigt das neu gebaute lokale Brave-Image.

## Block 4 — adaptives Compose-Profil bereitstellen

Für den gemeinsamen Checkpoint muss **vor diesem Block** der frische Token aus dem Abschnitt „Frischen Token ohne Host-OpenSSL erzeugen“ in [`VIEW_ONLY_SHARING.md`](VIEW_ONLY_SHARING.md) in `.env` eingetragen sein. Dieser Block gibt keine aufgelösten Umgebungswerte aus.

```bash
(
  set -e
  docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml config --quiet
  docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml up -d --force-recreate
  docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml ps
  NEKO_CONTAINER_ID="$(docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml ps -q neko)"
  docker inspect --format 'container={{.Name}} image={{.Config.Image}} image_id={{.Image}} started={{.State.StartedAt}} status={{.State.Status}}' "$NEKO_CONTAINER_ID"
  docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml logs --since=5m --no-color neko
) 2>&1 | tee "$NEKO_RESULT_DIR/04-deployment.txt"
```

Vor den Gerätetests muss Folgendes stimmen:

- `neko` läuft ohne Neustartschleife;
- das erwartete lokale Brave-Image wird verwendet;
- kein Konfigurations-, GStreamer- oder Pluginfehler;
- Profil, Downloads und Brave-Policy sind weiterhin vorhanden;
- Admin- und Mitgliederlogin, Audio, Video und Kontrolle funktionieren;
- das Policy-Ziel bleibt `/etc/brave/policies/managed/policies.json`.

## Metrik-Snapshot ohne Host-Werkzeuge

Der Hilfsservice verwendet auf dem Linux-Host das Hostnetz und liest standardmäßig `http://127.0.0.1:8082/metrics`:

```bash
docker compose -f docker-compose.validation.yaml run --rm metrics-snapshot
```

Wenn `NEKO_HTTP_PORT` nicht `8082` ist, vor dem Aufruf beispielsweise setzen:

```bash
export NEKO_METRICS_URL="http://127.0.0.1:ANDERER_PORT/metrics"
```

Die Ausgabe enthält nur die für Sessionzustand, Tierwahl, Bitrate und Peer-Drops relevanten Metriken, keine Compose-Geheimnisse.

## Vor den iPhone-Phasen erfassen

Benötigt werden drei gleichzeitig aktive Zuschauer mit sichtbar wechselndem Desktopinhalt:

- `H1`: gesunder Desktop;
- `H2`: gesunder iPad-/zweiter Zuschauer;
- `C`: das zu prüfende iPhone.

Manuell notieren:

- Commit und Image-ID aus Block 0/4;
- iPhone-Modell, iOS- und Safari-Version;
- Verfügbarkeit des Safari Web Inspectors;
- Session-IDs von H1, H2 und C;
- anfänglicher Muted-/Playing-Zustand;
- ob beim ersten Join ein Play-Tap erforderlich war.

Baseline-Metriken speichern:

```bash
docker compose -f docker-compose.validation.yaml run --rm metrics-snapshot 2>&1 | tee "$NEKO_RESULT_DIR/10-ios-baseline.txt"
```

Im Safari Web Inspector nach Möglichkeit die Clientmeldungen mit diesen Texten erhalten:

```text
peer ice connection state changed
ICE recovery timeout
scheduling application reconnect attempt
starting application reconnect attempt
connected
Autoplay blocked
```

Ohne Web Inspector muss eine alternative Aufzeichnung die Anzahl paralleler/gestarteter Versuche eindeutig belegen. Reine Sichtbeobachtung genügt für die Vier-Versuche-Grenze nicht.

## Phase A — bestehender Peer erholt sich

1. H1, dann H2, dann C verbinden und 60 Sekunden mit funktionierendem Audio/Video warten.
2. Auf C Flugmodus für **5 Sekunden** einschalten, danach ausschalten.
3. Weder Seite neu laden noch Login oder Play betätigen.
4. 60 Sekunden beobachten.
5. Erst wenn Media zurückgekehrt ist, einen eventuell sichtbaren zentralen Play-Button genau einmal betätigen und separat als Safari-Policy-Ergebnis notieren.

Danach ausführen:

```bash
docker compose -f docker-compose.validation.yaml run --rm metrics-snapshot 2>&1 | tee "$NEKO_RESULT_DIR/11-ios-phase-a.txt"
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml logs --since=10m --no-color neko 2>&1 | tee "$NEKO_RESULT_DIR/12-ios-phase-a-server.log"
```

Phase A besteht nur, wenn:

- C dieselbe Server-Session-ID behält;
- ICE von `disconnected` nach `connected`/`completed` zurückkehrt;
- kein `starting application reconnect attempt` erscheint;
- H1/H2 nicht reconnecten, nicht stocken, auf `high` bleiben und keine neuen Video-Drops erhalten;
- kein Reload verwendet wurde.

## Phase B — Ersatzsession ohne Reload

1. Alle drei Zuschauer wieder vollständig stabilisieren.
2. Auf C Flugmodus für **15 Sekunden** einschalten. Falls der alte Peer dadurch nicht geschlossen wird und auch der ICE-Timeout nicht auslöst, nur diese Unterbrechung verlängern, bis eine dieser Grenzen nachweislich erreicht ist; tatsächliche Dauer notieren.
3. Flugmodus ausschalten. Nicht neu laden, nicht navigieren, Login nicht erneut absenden und Play zunächst nicht drücken.
4. Bis zu 90 Sekunden beobachten.
5. Nach vorhandenem Media einen eventuell notwendigen zentralen Play-Button einmal drücken und als Autoplay-Ergebnis notieren.

Danach ausführen:

```bash
docker compose -f docker-compose.validation.yaml run --rm metrics-snapshot 2>&1 | tee "$NEKO_RESULT_DIR/13-ios-phase-b.txt"
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml logs --since=15m --no-color neko 2>&1 | tee "$NEKO_RESULT_DIR/14-ios-phase-b-server.log"
```

Phase B besteht nur, wenn:

- immer höchstens ein Reconnectversuch gleichzeitig läuft;
- höchstens vier Versuche gestartet werden;
- eine neue C-Session-ID entsteht und Audio/Video ohne Reload oder manuellen Login zurückkehren;
- alte Session-IDs verschwinden oder inaktiv werden;
- kein veralteter Callback die Ersatzsession wieder trennt;
- H1/H2 durchgehend verwendbar bleiben, auf `high` bleiben und keine neuen Peer-Video-Drops erhalten.

## Phase C — begrenztes Ausschöpfen und bewusster Disconnect

1. C wieder gesund verbinden.
2. C so lange offline halten, bis alle automatischen Versuche beendet sind. Wegen des 15-Sekunden-Verbindungstimeouts pro Versuch bis zu **90 Sekunden** warten.
3. Prüfen, dass höchstens vier Meldungen `starting application reconnect attempt` vorkommen und danach keine weitere Aktivität folgt.
4. Netz wiederherstellen. Das ausgefüllte Loginformular manuell absenden; kein Reload.
5. Nach erfolgreicher Verbindung C durch einen Admin kicken oder einen anderen absichtlichen `system/disconnect` auslösen.
6. Prüfen, dass danach kein automatischer Reconnect beginnt.

Danach ausführen:

```bash
docker compose -f docker-compose.validation.yaml run --rm metrics-snapshot 2>&1 | tee "$NEKO_RESULT_DIR/15-ios-phase-c.txt"
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml logs --since=30m --no-color neko 2>&1 | tee "$NEKO_RESULT_DIR/16-ios-phase-c-server.log"
```

## Gemeinsame Regression nach der Wiederherstellung

Nach Phase C auf C sowie H1/H2 prüfen:

- Touch und Trackpad;
- mobile Tastatur und Keyboard-Helper;
- Orientierung und Vollbild;
- Play, Mute/Unmute, Lautstärke und Audio;
- normale Request/Grant/Release-/Admin-Kontrollsemantik;
- Dateiübertragung mit den konfigurierten Mitglieds-/Adminrechten;
- keine neue Störung der beiden gesunden Zuschauer.

## Antwortvorlage für die iOS-Evidenz

Zusammen mit den Dateien `00` bis `16` diese ausgefüllte Kurzfassung übergeben:

```text
IOS CHECKPOINT
Commit:
Image-ID:
Profil: adaptive / base
iPhone / iOS / Safari:
Web Inspector verfügbar: ja / nein
H1 / H2 / C Session-IDs vor Start:

Phase A: PASS / FAIL
C Session-ID vorher/nachher:
ICE-Zustände:
Application-Reconnectversuche:
Play-Tap nötig: ja / nein
H1/H2 Beobachtung und Drop-Deltas:

Phase B: PASS / FAIL
Unterbrechungsdauer:
C Session-ID vorher/nachher:
Gestartete Versuche und maximale Parallelität:
Zeit bis Media zurück:
Play-Tap nötig: ja / nein
H1/H2 Beobachtung und Drop-Deltas:

Phase C: PASS / FAIL
Gesamtzahl gestarteter Versuche:
Manueller Login ohne Reload erfolgreich: ja / nein
Reconnect nach Admin-Kick/system disconnect: ja / nein

Touch/Trackpad/Keyboard/Orientierung/Vollbild/Audio/Kontrolle/Dateien:
Abweichungen oder Fehler:
```

Der Block ist erst zielserververifiziert, wenn alle Phasen und Regressionspunkte belegt sind. Bis dahin bleibt die Aussage: **implementiert und statisch geprüft; No-Reload-Verhalten auf dem Zielserver ausstehend**.

## Rollback

Beim adaptiven Profil das gesicherte Image ohne Volume-Löschung starten:

Falls inzwischen eine neue Shell geöffnet wurde, `NEKO_ROLLBACK_IMAGE` zuerst auf den in `00-provenance.txt` protokollierten Wert setzen.

```bash
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml down
NEKO_IMAGE="$NEKO_ROLLBACK_IMAGE" \
  docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml up -d --force-recreate
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml ps
```

Kein `down -v` verwenden. Die Recovery-Änderung benötigt keine Datenmigration. Nach einem Rollback Login, Video, Audio, Kontrolle, Profil und Policy erneut kurz prüfen.
