export const logout = 'Logga ut'
export const unsupported = 'Den här webbläsaren stöder inte WebRTC'
export const admin_loggedin = 'Du är inloggad som administratör'
export const you = 'Du'
export const somebody = 'Någon'
export const send_a_message = 'Skicka ett meddelande'

export const side = {
  chat: 'Chatt',
  files: 'Filer',
  settings: 'Inställningar',
  users: 'Användare',
}

export const connect = {
  login_title: 'Logga in',
  invitation_title: 'Du har blivit inbjuden till detta rum',
  displayname: 'Ange ditt visningsnamn',
  password: 'Lösenord',
  connect: 'Anslut',
  error: 'Inloggningsfel',
  empty_displayname: 'Visningsnamnet kan inte vara tomt.',
}

export const context = {
  ignore: 'Ignorera',
  unignore: 'Sluta ignorera',
  mute: 'Tysta',
  unmute: 'Avtysta',
  release: 'Tvinga frigöring av kontroller',
  take: 'Tvinga ta kontroller',
  give: 'Ge kontroller',
  kick: 'Sparka',
  ban: 'Bannlys IP',
  confirm: {
    kick_title: 'Sparka {name}?',
    kick_text: 'Är du säker på att du vill sparka {name}?',
    ban_title: 'Bannlysa {name}?',
    ban_text: 'Vill du bannlysa {name}? Du måste starta om servern för att ångra detta.',
    mute_title: 'Tysta {name}?',
    mute_text: 'Är du säker på att du vill tysta {name}?',
    unmute_title: 'Avtysta {name}?',
    unmute_text: 'Vill du avtysta {name}?',
    button_yes: 'Ja',
    button_cancel: 'Avbryt',
  },
}

export const controls = {
  release: 'Frigör kontroller',
  request: 'Begär kontroller',
  lock: 'Lås kontroller',
  unlock: 'Lås upp kontroller',
  has: 'Du har kontroll',
  hasnot: 'Du har inte kontroll',
}

export const locks = {
  control: {
    lock: 'Lås kontroller (för användare)',
    unlock: 'Lås upp kontroller (för användare)',
    locked: 'Kontroller låsta (för användare)',
    unlocked: 'Kontroller upplåsta (för användare)',
    notif_locked: 'låste kontroller för användare',
    notif_unlocked: 'låste upp kontroller för användare',
  },
  login: {
    lock: 'Lås rum (för användare)',
    unlock: 'Lås upp rum (för användare)',
    locked: 'Rum låst (för användare)',
    unlocked: 'Rum upplåst (för användare)',
    notif_locked: 'låste rummet',
    notif_unlocked: 'låste upp rummet',
  },
  file_transfer: {
    lock: 'Lås överföring (för användare)',
    unlock: 'Lås upp överföring (för användare)',
    locked: 'Överföring låst (för användare)',
    unlocked: 'Överföring upplåst (för användare)',
    notif_locked: 'låste filöverföring',
    notif_unlocked: 'låste upp filöverföring',
  },
}

export const setting = {
  scroll: 'Scrollkänslighet',
  scroll_invert: 'Invertera scroll',
  autoplay: 'Spela upp video automatiskt',
  ignore_emotes: 'Ignorera emotes',
  chat_sound: 'Spela chattljud',
  keyboard_layout: 'Tangentbordslayout',
  broadcast_title: 'Live-sändning',
}

export const connection = {
  logged_out: 'Du har loggats ut.',
  reconnecting: 'Återansluter...',
  connected: 'Ansluten',
  disconnected: 'Frånkopplad',
  kicked: 'Du har tagits bort från detta rum.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} anslöt',
  disconnected: '{name} frånkopplades',
  controls_taken: '{name} tog kontrollerna',
  controls_taken_force: 'tog kontrollerna med tvång',
  controls_taken_steal: 'tog kontrollerna från {name}',
  controls_released: '{name} släppte kontrollerna',
  controls_released_force: 'släppte kontrollerna med tvång',
  controls_released_steal: 'släppte kontrollerna från {name}',
  controls_given: 'gav kontrollerna till {name}',
  controls_has: '{name} har kontrollerna',
  controls_has_alt: 'Men jag lät personen veta att du ville ha dem',
  controls_requesting: '{name} begär kontrollerna',
  resolution: 'ändrade upplösningen till {width}x{height}@{rate}',
  banned: 'bannlyste {name}',
  kicked: 'sparkade {name}',
  muted: 'tystade {name}',
  unmuted: 'avtystade {name}',
}

export const files = {
  downloads: 'Nedladdningar',
  uploads: 'Uppladdningar',
  upload_here: 'Klicka eller dra filer hit för att ladda upp',
}
