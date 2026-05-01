export const logout = 'Logg ut'
export const unsupported = 'Denne nettleseren støtter ikke WebRTC'
export const admin_loggedin = 'Du er innlogget som admin'
export const you = 'Du'
export const somebody = 'Noen'
export const send_a_message = 'Send en melding'

export const side = {
  chat: 'Chat',
  files: 'Filer',
  settings: 'Innstillinger',
  users: 'Brukere',
}

export const connect = {
  login_title: 'Vennligst logg inn',
  invitation_title: 'Du har blitt invitert til dette rommet',
  displayname: 'Skriv inn visningsnavnet ditt',
  password: 'Passord',
  connect: 'Koble til',
  error: 'Innloggingsfeil',
  empty_displayname: 'Visningsnavn kan ikke være tomt.',
}

export const context = {
  ignore: 'Ignorer',
  unignore: 'Slutt å ignorere',
  mute: 'Demp',
  unmute: 'Opphev demping',
  release: 'Tvangsfrigi kontroller',
  take: 'Tvangsoverta kontroller',
  give: 'Gi kontroller',
  kick: 'Spark',
  ban: 'Bannlys IP',
  confirm: {
    kick_title: 'Sparke {name}?',
    kick_text: 'Er du sikker på at du vil sparke {name}?',
    ban_title: 'Bannlyse {name}?',
    ban_text: 'Vil du bannlyse {name}? Du må starte serveren på nytt for å angre.',
    mute_title: 'Dempe {name}?',
    mute_text: 'Er du sikker på at du vil dempe {name}?',
    unmute_title: 'Oppheve demping av {name}?',
    unmute_text: 'Vil du oppheve dempingen av {name}?',
    button_yes: 'Ja',
    button_cancel: 'Avbryt',
  },
}

export const controls = {
  release: 'Frigi kontroller',
  request: 'Be om kontroller',
  lock: 'Lås kontroller',
  unlock: 'Lås opp kontroller',
  has: 'Du har kontroll',
  hasnot: 'Du har ikke kontroll',
  mic_on: 'Aktiver mikrofon',
  mic_off: 'Deaktiver mikrofon',
  mic_error: 'Mikrofonfeil',
}

export const locks = {
  control: {
    lock: 'Lås kontroller (for brukere)',
    unlock: 'Lås opp kontroller (for brukere)',
    locked: 'Kontroller låst (for brukere)',
    unlocked: 'Kontroller låst opp (for brukere)',
    notif_locked: 'låste kontroller for brukere',
    notif_unlocked: 'låste opp kontroller for brukere',
  },
  login: {
    lock: 'Lås rom (for brukere)',
    unlock: 'Lås opp rom (for brukere)',
    locked: 'Rom låst (for brukere)',
    unlocked: 'Rom låst opp (for brukere)',
    notif_locked: 'låste rommet',
    notif_unlocked: 'låste opp rommet',
  },
  file_transfer: {
    lock: 'Lås filoverføring (for brukere)',
    unlock: 'Lås opp filoverføring (for brukere)',
    locked: 'Filoverføring låst (for brukere)',
    unlocked: 'Filoverføring låst opp (for brukere)',
    notif_locked: 'låste filoverføring',
    notif_unlocked: 'låste opp filoverføring',
  },
}

export const setting = {
  scroll: 'Rullingssensitivitet',
  scroll_invert: 'Inverter rulling',
  trackpad_mode: 'Styreflatemodus',
  trackpad_mode_description: 'Flytt markøren relativt som en styreflate',
  trackpad_mode_mobile_hint: 'To fingre for å rulle, trykk for å klikke, hold for høyreklikk',
  mobile: 'Mobil',
  desktop_only_inactive: 'Kun berøring',
  autoplay: 'Spill video automatisk',
  ignore_emotes: 'Ignorer smilefjes',
  chat_sound: 'Sludringslyd',
  keyboard_layout: 'Tastaturoppsett',
  broadcast_title: 'Direktesending',
}

export const connection = {
  logged_out: 'Du har blitt utlogget.',
  reconnecting: 'Kobler til igjen...',
  connected: 'Tilkoblet',
  disconnected: 'Frakoblet',
  kicked: 'Du har blitt fjernet fra dette rommet.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} koblet til',
  disconnected: '{name} koblet fra',
  controls_taken: '{name} tok kontrollene',
  controls_taken_force: 'tok kontrollene med tvang',
  controls_taken_steal: 'tok kontrollene fra {name}',
  controls_released: '{name} frigav kontrollene',
  controls_released_force: 'frigav kontrollene med tvang',
  controls_released_steal: 'frigav kontrollene fra {name}',
  controls_given: 'ga kontrollene til {name}',
  controls_has: '{name} har kontrollene',
  controls_has_alt: 'Men jeg informerte personen om at du ønsket det',
  controls_requesting: '{name} ber om kontrollene',
  resolution: 'endret oppløsningen til {width}x{height}@{rate}',
  banned: 'bannlyste {name}',
  kicked: 'sparket {name}',
  muted: 'dempet {name}',
  unmuted: 'opphevet dempingen av {name}',
}

export const files = {
  downloads: 'Nedlastinger',
  uploads: 'Opplastinger',
  upload_here: 'Klikk eller dra filer hit for å laste opp',
}
