# Serverseitig erzwungenes View-only-Sharing

Status: **auf `testing` implementiert und statisch geprüft; Docker-Builds, fokussierte Servertests und die Drei-Rollen-Laufzeitmatrix auf dem Zielserver stehen noch aus**.

Dieses Runbook ist auf einen Linux-Zielserver ohne lokal installiertes Node.js, npm, Go oder Python zugeschnitten. Die automatisierbaren Prüfungen laufen über [`docker-compose.validation.yaml`](../docker-compose.validation.yaml). Nur die tatsächlichen Browser-, Rollen-, Eingabe-, Mikrofon- und Widerrufstests bleiben manuell.

Der Checkpoint muss gemeinsam mit [`IOS_RECOVERY.md`](IOS_RECOVERY.md) auf demselben exakten `testing`-Commit und demselben Neko-Image erfolgen.

## Sicherheitsmodell in Kurzform

Der optionale Token erzeugt eine passive Session im selben Neko-Raum und mit demselben WebRTC-Audio/-Video. Die Session darf weder Steuerung noch Maus, Tastatur, Touch, Zwischenablage, Dateioperationen, Mikrofon/Kamera, Chat-Senden, Plugins oder Adminfunktionen verwenden.

Die Autorisierung beruht auf `MemberProfile.IsViewOnly` und serverseitigen Eingangsprüfungen, nicht auf ausgeblendeten UI-Elementen und nicht auf WebRTC. Geprüft werden:

- Profilnormalisierung und Hostzuweisung;
- authentifizierte HTTP-Routen;
- aktuelles und Legacy-WebSocket-Protokoll;
- moderner und Legacy-WebRTC-Data-Channel;
- eingehende Media-Tracks;
- Chat, Dateiübertragung und Open-in-App;
- Sessionpersistenz und Widerruf.

Die erste Umsetzung verwendet WebRTC weiterhin als Empfangspfad. Der passive Marker ist absichtlich transportneutral, damit ein späterer HLS-/LL-HLS-Empfänger dieselbe Autorisierungsgrenze verwenden kann.

## Token, URL und Geheimhaltung

`member.multiuser.view_only_token` ist leer oder enthält genau 64 hexadezimale Zeichen, also 256 Bit. Die URL lautet:

```text
https://neko.example/#/watch/<64-hex-token>
```

Das Fragment gelangt nicht in den normalen HTTP-Request-Pfad, die Query oder den `Referer`. Der Client überträgt den Token als WebSocket-Subprotokoll `neko-view.<token>`. Der Token kann trotzdem in Browserverlauf, Bookmarks, Screenshots und kopierten Links verbleiben. Daher:

- Internetzugriff nur über TLS;
- jeden Empfänger und jedes Gerät mit dem Link als Tokeninhaber behandeln;
- Reverse Proxies dürfen `Sec-WebSocket-Protocol` nicht protokollieren;
- den Token niemals in Git, Chat-Output, Screenshots oder Ergebnisdateien aufnehmen;
- Mitglieds-, Admin- und View-only-Credential müssen verschieden sein.

Ein leerer Wert deaktiviert den Link. Ein syntaktisch falscher Subprotokollwert erhält HTTP 400; ein formal gültiger, aber falscher oder deaktivierter Token endet mit dem normalen Authentifizierungsfehler.

## Lebensdauer und Widerruf

Der Token hat keine Uhrzeit-basierte Ablaufzeit. Er gilt, bis er ersetzt oder entfernt und der Neko-Service neu erstellt wird. View-only-Sessions werden nicht in `session.file` gespeichert; alte serialisierte passive Sessions werden beim Laden ignoriert.

Der harte Widerruf besteht deshalb immer aus:

1. Token in `.env` ersetzen oder leeren;
2. `docker compose ... config --quiet` ausführen;
3. Neko-Service mit `--force-recreate` neu erstellen;
4. neuen Link nur bei weiterhin gewünschtem Sharing verteilen.

## Reihenfolge für den gemeinsamen Checkpoint

1. [`IOS_RECOVERY.md`](IOS_RECOVERY.md), Blöcke 0–3 ausführen. Dadurch entstehen Provenienz, containerisierte Client-/Serverchecks, ein Rollback-Tag und die neuen Images.
2. Den nachfolgenden frischen View-only-Token setzen.
3. [`IOS_RECOVERY.md`](IOS_RECOVERY.md), Block 4 ausführen. Dadurch wird das adaptive Profil mit dem neuen Token bereitgestellt.
4. Die automatisierte HTTP-Prüfung und anschließend die manuelle Drei-Rollen-Matrix dieses Dokuments ausführen.
5. Danach die iPhone-Phasen A–C aus `IOS_RECOVERY.md` auf exakt derselben Bereitstellung ausführen.

Die Variablen `NEKO_VALIDATION_COMMIT`, `NEKO_RESULT_DIR` und `NEKO_ROLLBACK_IMAGE` aus iOS-Block 0 müssen in derselben Shell weiter vorhanden sein.

## Frischen Token ohne Host-OpenSSL erzeugen

Die Ausgabe dieses Befehls ist das Geheimnis. Sie darf **nicht** mit `tee` aufgezeichnet und später nicht mit den übrigen Outputs übergeben werden.

```bash
docker compose -f docker-compose.validation.yaml pull token-generator
docker compose -f docker-compose.validation.yaml run --rm token-generator
```

Den ausgegebenen 64-Zeichen-Wert im Editor als `NEKO_VIEW_ONLY_TOKEN=...` in der ignorierten `.env` eintragen. Danach nur die stille Konfigurationsprüfung verwenden:

```bash
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml config --quiet
```

Anschließend iOS-Block 4 ausführen. Niemals die normale Ausgabe von `docker compose config` speichern oder weitergeben, weil sie Credential-Werte expandiert.

## Automatisierte HTTP-Grenzprüfung in Docker Compose

Nach erfolgreicher Bereitstellung prüft dieser Container:

- View-only-Login und normalisiertes passives Profil;
- erlaubte `whoami`-, Statistik- und Screen-Lesezugriffe;
- HTTP 403 für Profil-, Session-, Mitglieder-, Kontrolle-, Tastatur-, Zwischenablage-, Upload-, Chat-, Datei- und Open-in-App-Zugriffe;
- anschließendes sauberes Logout.

Der Share-Token wird aus `.env` in den kurzlebigen Container gegeben, aber niemals ausgegeben. Bei einem abweichenden HTTP-Port oder Pfad zuerst `NEKO_BASE_URL` setzen.

```bash
export NEKO_BASE_URL="http://127.0.0.1:8082"

(
  set -e
  docker compose -f docker-compose.validation.yaml pull view-only-http-checks
  docker compose -f docker-compose.validation.yaml run --rm view-only-http-checks
) 2>&1 | tee "$NEKO_RESULT_DIR/20-view-only-http.txt"
```

Erwartung: jede Zeile beginnt mit `PASS`, zuletzt erscheint `PASS all view-only HTTP boundary checks`. Ein `FAIL`, ein unerwarteter Status oder ein fehlender Sessiontoken beendet diesen Teil als Fehler. Cookie-Authentifizierung muss für diesen Test wie in der aktuellen Deployment-Baseline deaktiviert sein.

Diese Prüfung ersetzt nicht die WebSocket-, Data-Channel-, Media- und Rollenmatrix.

## Drei Teilnehmer vorbereiten

Drei gleichzeitige Teilnehmer im selben Desktop verwenden:

- `M`: normales Mitglied;
- `A`: Administrator;
- `V`: View-only-Link.

Vor dem Start notieren:

- Commit, Image-ID und `adaptive`/`base`;
- Browser/Gerät jedes Teilnehmers;
- Session-ID von M, A und V;
- ob Chat, Dateiübertragung und Open-in-App in dieser Bereitstellung aktiviert sind.

Baseline-Metriken speichern:

```bash
docker compose -f docker-compose.validation.yaml run --rm metrics-snapshot 2>&1 | tee "$NEKO_RESULT_DIR/21-view-only-baseline.txt"
```

## Sollmatrix

| Prüfung | Mitglied `M` | Admin `A` | View-only `V` |
| --- | --- | --- | --- |
| gleicher Raum, Audio und Video | erlaubt | erlaubt | erlaubt |
| Play/Mute/Lautstärke/Vollbild/PiP | erlaubt | erlaubt | erlaubt |
| Kontrolle anfordern oder erhalten | gemäß Raumregeln | erlaubt | verweigert |
| Maus/Scroll/Tasten/Tastatur/Touch | nur Host | Host-/Adminregeln | verweigert |
| Zwischenablage lesen/schreiben | erlaubter Host | erlaubter Host | verweigert |
| Datei-Liste/Download/Upload/Löschen | konfigurierte Rechte | Adminrechte | verweigert; keine Dateinamen |
| Mikrofon/Kamera veröffentlichen | Profil-/Kontrollregeln | Profil-/Kontrollregeln | gestoppt |
| Chat oder beliebiges Plugin-Event senden | konfigurierte Rechte | konfigurierte Rechte | verweigert |
| Profil-/Session-/Member-/Plugin-HTTP-Mutation | Capability-Regeln | soweit Admin erlaubt | HTTP 403 |
| Lock/Take/Give/Revoke | verweigert | erlaubt | verweigert |

## Phase 1 — normale Mitglieds-/Adminsemantik

1. M verbindet sich und fordert Kontrolle an.
2. M gibt Kontrolle wieder frei.
3. A übernimmt Kontrolle, gibt sie an M und entzieht sie wieder.
4. A aktiviert/deaktiviert die Kontrollsperre und bestätigt die bisherige Lock-Semantik.
5. Audio, Video, Touch, Zwischenablage sowie die konfigurierten Dateioperationen für M/A kurz prüfen.

Ergebnis: M/A müssen sich exakt wie vor dem View-only-Block verhalten.

## Phase 2 — passive Oberfläche und gemeinsamer Raum

1. V über `#/watch/<token>` verbinden.
2. Prüfen, dass V denselben wechselnden Desktopinhalt und dasselbe Audio wie M/A erhält.
3. V darf nur Video-/Playback-Funktionen sehen: Play, Mute, Lautstärke, Vollbild und PiP, soweit der Browser sie unterstützt.
4. Raum-, Chat-, Datei-, Zwischenablage-, Eingabe-, Mikrofon- und Adminoberflächen dürfen nicht sichtbar sein.
5. Falls WebSocket-Frames inspiziert werden, muss die Dateiübertragung für V deaktiviert sein und eine eventuelle Dateiliste leer bleiben; kein Dateiname darf V erreichen.

Die ausgeblendete Oberfläche ist nur UX-Evidenz; die folgenden serverseitigen Denial-Phasen bleiben verpflichtend.

## Phase 3 — Hostzuweisung verweigern

1. A versucht, V Kontrolle zu geben.
2. A versucht danach die normale Übergabe an M.
3. Prüfen, dass V niemals Host wird, keine Eingabe den Desktop erreicht und M weiterhin Kontrolle erhalten kann.
4. Erwartete Serverwarnung: die Hostzuweisung an eine nicht interaktive Session wird verweigert.

## Phase 4 — Legacy-WebSocket und Data Channel aus dem View-only-Browser

Im Entwicklerkonsolenfenster von V diesen Block einfügen. Er enthält und druckt keinen Token:

```javascript
$client.sendMessage('control/request')
$client.sendMessage('control/clipboard', { text: 'view-only-denial-check' })
$client.sendMessage('control/touchbegin', { id: 1, x: 100, y: 100, pressure: 1 })
$client.sendMessage('chat/message', { content: 'view-only-denial-check' })
$client.sendMessage('filetransfer/refresh')
$client.sendMessage('openinapp/openlink', { text: 'https://example.invalid' })
$client.sendMessage('admin/control')
$client.sendData('mousemove', { x: 100, y: 100 })
$client.sendData('wheel', { x: 0, y: 120 })
$client.sendData('keydown', { key: 0x41 })
$client.sendData('mousedown', { key: 1 })
```

Danach prüfen:

- keinerlei Maus-, Scroll-, Tastatur-, Klick-, Chat-, Datei-, Link- oder Kontrolleffekt;
- V bleibt verbunden und kann weitersehen;
- Serverlogs enthalten Denial-Warnungen für Legacy-WebSocket bzw. Legacy-Data-Channel.

Logs speichern:

```bash
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml logs --since=10m --no-color neko 2>&1 | tee "$NEKO_RESULT_DIR/22-view-only-legacy-input.log"
```

## Phase 5 — aktuelles WebSocket-Protokoll direkt prüfen

Der normale Fork-Client verwendet den Legacy-Adapter. Damit zusätzlich der aktuelle WebSocket-Dispatcher der laufenden Bereitstellung geprüft wird, diesen vollständigen Block in der Entwicklerkonsole von V ausführen. Er liest den Share-Token nur aus dem Fragment, gibt ihn nicht aus, erstellt eine separate kurzlebige View-only-Session und loggt sie anschließend wieder aus.

```javascript
(async () => {
  const shareToken = location.hash.match(/^#\/watch\/([0-9a-fA-F]{64})$/)?.[1]
  if (!shareToken) throw new Error('view-only token missing from fragment')

  const prefix = location.pathname
    .replace(/\/(index|login)\.html?$/i, '/')
    .replace(/\/$/, '')
  const api = `${location.origin}${prefix}/api`
  const wsScheme = location.protocol === 'https:' ? 'wss' : 'ws'

  const response = await fetch(`${api}/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: `ws-check-${Date.now()}`, password: shareToken }),
  })
  if (!response.ok) throw new Error(`login failed: HTTP ${response.status}`)
  const session = await response.json()
  if (!session.profile?.is_view_only) throw new Error('session is not view-only')

  const socket = new WebSocket(
    `${wsScheme}://${location.host}${prefix}/api/ws?token=${encodeURIComponent(session.token)}`,
  )
  await new Promise((resolve, reject) => {
    socket.onopen = resolve
    socket.onerror = () => reject(new Error('current websocket failed to open'))
  })

  for (const event of [
    'control/request',
    'control/touchbegin',
    'clipboard/set',
    'screen/set',
    'send/broadcast',
    'chat/message',
    'plugin/future-interactive-event',
  ]) {
    socket.send(JSON.stringify({ event, payload: {} }))
  }

  await new Promise((resolve) => setTimeout(resolve, 1500))
  socket.close()
  await fetch(`${api}/logout`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${session.token}` },
  })
  console.log(`PASS current websocket denial probe; session ${session.id}`)
})()
```

Danach muss die Konsole `PASS current websocket denial probe` melden. Der Desktop darf sich nicht ändern; V/M/A dürfen keinen Chat- oder Plugin-Effekt sehen. Serverlogs müssen die verweigerten aktuellen WebSocket-Ereignisse zeigen:

```bash
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml logs --since=10m --no-color neko 2>&1 | tee "$NEKO_RESULT_DIR/23-view-only-current-ws.log"
```

Die containerisierten Go-Tests aus iOS-Block 2 prüfen zusätzlich die vollständigen aktuellen/Legacy-Allowlisten und beide Data-Channel-Handler am exakten Commit. Falls die Zielbereitstellung später einen nativen aktuellen Client verwendet, muss dessen echter Data Channel zusätzlich adversarial geprüft werden; der jetzige Legacy-Client deckt zur Laufzeit den tatsächlich verwendeten Legacy-Kanal ab.

## Phase 6 — eingehendes Mikrofon/Media verweigern

In der Entwicklerkonsole von V ausführen und eine eventuelle Browser-Mikrofonabfrage nur für diesen Test erlauben:

```javascript
await $client.enableMicrophone()
```

Erwartung:

- M und A hören niemals Audio von V;
- V erhält weder Media-Share- noch Kontrollrechte;
- der Server stoppt einen tatsächlich eingehenden Track und protokolliert, dass Media-Sharing für die Session deaktiviert ist.

Wenn der Browser bereits lokal keinen Track aushandelt, ist nur der Browserpfad fehlgeschlagen und der serverseitige Laufzeittest noch nicht bewiesen. Dies als Abweichung vermerken; der fokussierte Go-Test allein ersetzt keinen tatsächlich beim Server eingehenden Track.

```bash
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml logs --since=10m --no-color neko 2>&1 | tee "$NEKO_RESULT_DIR/24-view-only-media.log"
```

## Phase 7 — Refresh und kurzzeitige Unterbrechung von V

1. V neu laden und danach kurz offline/online schalten.
2. V muss erneut ausschließlich view-only verbunden sein.
3. M/A dürfen nicht reconnecten, stocken oder Kontrolle verlieren.
4. Im adaptiven Profil bleiben gesunde Zuschauer auf `high` ohne neue Peer-Video-Drops.

```bash
docker compose -f docker-compose.validation.yaml run --rm metrics-snapshot 2>&1 | tee "$NEKO_RESULT_DIR/25-view-only-reconnect.txt"
```

## Phase 8 — Tokenrotation und aktiven Widerruf beweisen

1. Mit noch verbundenem V einen neuen Token über den `token-generator` erzeugen. Die Ausgabe wieder nicht speichern oder weitergeben.
2. Den Wert in `.env` ersetzen.
3. Service mit demselben adaptiven Profil neu erstellen:

```bash
(
  set -e
  docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml config --quiet
  docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml up -d --force-recreate
  docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml ps
) 2>&1 | tee "$NEKO_RESULT_DIR/26-view-only-rotation.txt"
```

4. Prüfen: die alte aktive V-Session ist getrennt; der alte Link kann sich nicht anmelden; der neue Link kann sehen; Mitglieds- und Adminpasswort funktionieren unverändert.
5. Die automatisierte HTTP-Grenzprüfung mit dem neuen `.env`-Token wiederholen:

```bash
docker compose -f docker-compose.validation.yaml run --rm view-only-http-checks 2>&1 | tee "$NEKO_RESULT_DIR/27-view-only-http-after-rotation.txt"
```

## Phase 9 — optional wieder deaktivieren

Soll Sharing nach dem Checkpoint deaktiviert bleiben, `NEKO_VIEW_ONLY_TOKEN=` in `.env` leeren und nochmals ausführen:

```bash
(
  set -e
  docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml config --quiet
  docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml up -d --force-recreate
  docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml ps
) 2>&1 | tee "$NEKO_RESULT_DIR/28-view-only-disabled.txt"
```

Danach muss der letzte Link ebenfalls abgewiesen werden, während M/A weiterhin normal funktionieren.

## Vollständige Akzeptanz

Der View-only-Block gilt nur dann als zielserververifiziert, wenn:

- M, A und V denselben logischen Raum/Desktop sehen;
- V Audio/Video empfängt und alle als verweigert markierten Aktionen serverseitig scheitern;
- HTTP-Prüfung, Legacy-/aktuelles WebSocket, Data Channel, Plugins und ein tatsächlich eingehender Media-Track keine interaktive Wirkung erzielen;
- V keine Dateinamen erhält;
- A V keine Kontrolle geben kann, M/A-Kontroll- und Lock-Semantik aber unverändert bleibt;
- Rotation plus Service-Neuerstellung alte Links und aktive View-only-Sessions widerruft;
- normales Mitglied, Admin, Mobile/Touch, Dateiübertragung, adaptive Qualität und das gesamte iOS-Recovery-Runbook bestehen.

Bis alle Punkte belegt sind, lautet der Status weiterhin: **implementiert und statisch geprüft; Zielservervalidierung ausstehend**.

## Antwortvorlage für die View-only-Evidenz

Zusammen mit den Dateien `20` bis `28` diese Kurzfassung übergeben:

```text
VIEW-ONLY CHECKPOINT
Commit:
Image-ID:
Rollback-Image:
Profil: adaptive / base
M Browser/Gerät/Session-ID:
A Browser/Gerät/Session-ID:
V Browser/Gerät/Session-ID:

HTTP-Containercheck: PASS / FAIL
Normale M/A Kontrolle und Locks: PASS / FAIL
V gleicher Desktop und Audio/Video: PASS / FAIL
V nur Video-/Playback-Oberfläche: PASS / FAIL
A kann V keine Kontrolle geben: PASS / FAIL
Legacy-WebSocket-Denials: PASS / FAIL
Legacy-Data-Channel-Denials: PASS / FAIL
Aktuelles WebSocket-Denial-Skript: PASS / FAIL
Eingehender Mikrofontrack tatsächlich serverseitig gestoppt: PASS / FAIL / NICHT ERREICHT
V erhält keine Dateinamen: PASS / FAIL
Refresh/Unterbrechung von V ohne Störung M/A: PASS / FAIL
Alter Link und aktive V-Session nach Rotation widerrufen: PASS / FAIL
Neuer Link funktioniert weiterhin nur view-only: PASS / FAIL
M/A nach Rotation unverändert: PASS / FAIL
Sharing am Ende: neuer Token aktiv / deaktiviert

Abweichungen, Fehlermeldungen oder nicht anwendbare Punkte:
```

Die erzeugten Text-/Logdateien können nach manueller Kontrolle auf sensible Hostdaten als ein Archiv übergeben werden:

```bash
tar -C "$(dirname "$NEKO_RESULT_DIR")" \
  -czf "${NEKO_RESULT_DIR}.tar.gz" \
  "$(basename "$NEKO_RESULT_DIR")"
echo "${NEKO_RESULT_DIR}.tar.gz"
```

Den View-only-Token und `.env` niemals in dieses Verzeichnis kopieren.

## Rollback

Zuerst `NEKO_VIEW_ONLY_TOKEN=` leeren und den aktuellen Service neu erstellen. Ist zusätzlich ein Code-/Image-Rollback nötig, das in iOS-Block 3 gesicherte Image verwenden:

Falls inzwischen eine neue Shell geöffnet wurde, `NEKO_ROLLBACK_IMAGE` zuerst auf den in `00-provenance.txt` protokollierten Wert setzen.

```bash
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml down
NEKO_IMAGE="$NEKO_ROLLBACK_IMAGE" \
  docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml up -d --force-recreate
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml ps
```

Kein `down -v` verwenden. Es gibt keine Datenbank- oder persistente Datenmigration. Nach dem Rollback normale Mitglieds-/Adminanmeldung, Video, Audio, Kontrolle, Profil und Policy erneut prüfen.
