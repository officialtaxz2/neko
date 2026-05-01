export const logout = 'Abmelden'
export const unsupported = 'Dieser Browser unterstützt WebRTC nicht'
export const admin_loggedin = 'Du bist als Admin eingeloggt'
export const you = 'Du'
export const somebody = 'Jemand'
export const send_a_message = 'Nachricht senden'

export const side = {
  chat: 'Chat',
  files: 'Dateien',
  settings: 'Einstellungen',
  users: 'Benutzer',
}

export const connect = {
  login_title: 'Bitte einloggen',
  invitation_title: 'Du wurdest in diesen Raum eingeladen',
  displayname: 'Gib deinen Anzeigenamen ein',
  password: 'Passwort',
  connect: 'Verbinden',
  error: 'Anmeldefehler',
  empty_displayname: 'Der Anzeigename darf nicht leer sein.',
}

export const context = {
  ignore: 'Ignorieren',
  unignore: 'Nicht mehr ignorieren',
  mute: 'Stummschalten',
  unmute: 'Stummschaltung aufheben',
  release: 'Steuerung erzwungen freigeben',
  take: 'Steuerung erzwungen übernehmen',
  give: 'Steuerung übergeben',
  kick: 'Rauswerfen',
  ban: 'IP sperren',
  confirm: {
    kick_title: '{name} rauswerfen?',
    kick_text: 'Bist du sicher, dass du {name} rauswerfen möchtest?',
    ban_title: '{name} sperren?',
    ban_text: 'Möchtest du {name} sperren? Du musst den Server neu starten, um dies rückgängig zu machen.',
    mute_title: '{name} stummschalten?',
    mute_text: 'Bist du sicher, dass du {name} stummschalten möchtest?',
    unmute_title: '{name} Stummschaltung aufheben?',
    unmute_text: 'Möchtest du die Stummschaltung von {name} aufheben?',
    button_yes: 'Ja',
    button_cancel: 'Abbrechen',
  },
}

export const controls = {
  release: 'Steuerung freigeben',
  request: 'Steuerung anfragen',
  lock: 'Steuerung sperren',
  unlock: 'Steuerung entsperren',
  has: 'Du hast die Steuerung',
  hasnot: 'Du hast keine Steuerung',
  mic_on: 'Mikrofon aktivieren',
  mic_off: 'Mikrofon deaktivieren',
  mic_error: 'Mikrofonfehler',
}

export const locks = {
  control: {
    lock: 'Steuerung sperren (für Benutzer)',
    unlock: 'Steuerung entsperren (für Benutzer)',
    locked: 'Steuerung gesperrt (für Benutzer)',
    unlocked: 'Steuerung entsperrt (für Benutzer)',
    notif_locked: 'hat die Steuerung für Benutzer gesperrt',
    notif_unlocked: 'hat die Steuerung für Benutzer entsperrt',
  },
  login: {
    lock: 'Raum sperren (für Benutzer)',
    unlock: 'Raum entsperren (für Benutzer)',
    locked: 'Raum gesperrt (für Benutzer)',
    unlocked: 'Raum entsperrt (für Benutzer)',
    notif_locked: 'hat den Raum gesperrt',
    notif_unlocked: 'hat den Raum entsperrt',
  },
  file_transfer: {
    lock: 'Dateiübertragung sperren (für Benutzer)',
    unlock: 'Dateiübertragung entsperren (für Benutzer)',
    locked: 'Dateiübertragung gesperrt (für Benutzer)',
    unlocked: 'Dateiübertragung entsperrt (für Benutzer)',
    notif_locked: 'hat die Dateiübertragung gesperrt',
    notif_unlocked: 'hat die Dateiübertragung entsperrt',
  },
}

export const setting = {
  scroll: 'Scroll-Empfindlichkeit',
  scroll_invert: 'Bildlauf umkehren',
  trackpad_mode: 'Trackpad-Modus',
  trackpad_mode_description: 'Cursor relativ bewegen wie ein Touchpad',
  trackpad_mode_mobile_hint: 'Zwei Finger zum Scrollen, Tippen zum Klicken, Halten für Rechtsklick',
  mobile: 'Mobil',
  desktop_only_inactive: 'Nur Touch',
  autoplay: 'Autoplay Video',
  ignore_emotes: 'Emotes ignorieren',
  chat_sound: 'Chat-Sound abspielen',
  keyboard_layout: 'Tastaturbelegung',
  broadcast_title: 'Live-Übertragung',
}

export const connection = {
  logged_out: 'Du wurdest ausgeloggt.',
  reconnecting: 'Erneut verbinden...',
  connected: 'Verbindet',
  disconnected: 'Getrennt',
  kicked: 'Du wurdest aus diesem Raum entfernt.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} hat sich verbunden',
  disconnected: '{name} hat die Verbindung getrennt',
  controls_taken: '{name} hat die Steuerung übernommen',
  controls_taken_force: 'hat die Steuerung erzwungen übernommen',
  controls_taken_steal: 'hat die Steuerung von {name} übernommen',
  controls_released: '{name} hat die Steuerung freigegeben',
  controls_released_force: 'hat die Steuerung erzwungen freigegeben',
  controls_released_steal: 'hat die Steuerung von {name} freigegeben',
  controls_given: 'hat die Steuerung an {name} übergeben',
  controls_has: '{name} hat die Steuerung',
  controls_has_alt: 'Aber ich habe der Person mitgeteilt, dass du sie wolltest',
  controls_requesting: '{name} fragt nach der Steuerung',
  resolution: 'hat die Auflösung auf {width}x{height}@{rate} geändert',
  banned: 'hat {name} gesperrt',
  kicked: 'hat {name} rausgeworfen',
  muted: 'hat {name} stummgeschaltet',
  unmuted: 'hat die Stummschaltung von {name} aufgehoben',
}

export const files = {
  downloads: 'Downloads',
  uploads: 'Uploads',
  upload_here: 'Hier klicken oder Dateien ziehen zum Hochladen',
}
