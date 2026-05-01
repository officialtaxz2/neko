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
  login_title: 'Vänligen logga in',
  invitation_title: 'Du har blivit inbjuden till detta rum',
  displayname: 'Ange ditt visningsnamn',
  password: 'Lösenord',
  connect: 'Anslut',
  error: 'Inloggningsfel',
  empty_displayname: 'Visningsnamnet får inte vara tomt.',
}

export const context = {
  ignore: 'Ignorera',
  unignore: 'Sluta ignorera',
  mute: 'Tysta',
  unmute: 'Sluta tysta',
  release: 'Tvinga frigivning av kontroller',
  take: 'Tvinga ta kontroller',
  give: 'Ge kontroller',
  kick: 'Kicka',
  ban: 'Bannlys IP',
  confirm: {
    kick_title: 'Kicka {name}?',
    kick_text: 'Är du säker på att du vill kicka {name}?',
    ban_title: 'Bannlys {name}?',
    ban_text: 'Vill du bannlysa {name}? Du måste starta om servern för att ångra detta.',
    mute_title: 'Tysta {name}?',
    mute_text: 'Är du säker på att du vill tysta {name}?',
    unmute_title: 'Sluta tysta {name}?',
    unmute_text: 'Vill du sluta tysta {name}?',
    button_yes: 'Ja',
    button_cancel: 'Avbryt',
  },
}

export const controls = {
  release: 'Frigiv kontroller',
  request: 'Begär kontroller',
  lock: 'Lås kontroller',
  unlock: 'Lås upp kontroller',
  has: 'Du har kontroll',
  hasnot: 'Du har inte kontroll',
  mic_on: 'Aktivera mikrofon',
  mic_off: 'Inaktivera mikrofon',
  mic_error: 'Mikrofonfel',
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
    lock: 'Lås filöverföring (för användare)',
    unlock: 'Lås upp filöverföring (för användare)',
    locked: 'Filöverföring låst (för användare)',
    unlocked: 'Filöverföring upplåst (för användare)',
    notif_locked: 'låste filöverföring',
    notif_unlocked: 'låste upp filöverföring',
  },
}

export const setting = {
  scroll: 'Scrollkänslighet',
  scroll_invert: 'Invertera scroll',
  trackpad_mode: 'Styrplattläge',
  trackpad_mode_description: 'Flytta markören relativt som en styrplatta',
  trackpad_mode_mobile_hint: 'Två fingrar för att scrolla, tryck för att klicka, håll för högerklick',
  mobile: 'Mobil',
  desktop_only_inactive: 'Endast touch',
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
  disconnected: '{name} kopplades från',
  controls_taken: '{name} tog kontrollerna',
  controls_taken_force: 'tog kontrollerna med tvång',
  controls_taken_steal: 'tog kontrollerna från {name}',
  controls_released: '{name} frigjorde kontrollerna',
  controls_released_force: 'frigjorde kontrollerna med tvång',
  controls_released_steal: 'frigjorde kontrollerna från {name}',
  controls_given: 'gav kontrollerna till {name}',
  controls_has: '{name} har kontrollerna',
  controls_has_alt: 'Men jag informerade personen om att du ville ha det',
  controls_requesting: '{name} begär kontrollerna',
  resolution: 'ändrade upplösningen till {width}x{height}@{rate}',
  banned: 'bannlyste {name}',
  kicked: 'kickade {name}',
  muted: 'tystade {name}',
  unmuted: 'slutade tysta {name}',
}

export const files = {
  downloads: 'Nedladdningar',
  uploads: 'Uppladdningar',
  upload_here: 'Klicka eller dra filer hit för att ladda upp',
}
