export const logout = 'Wyloguj'
export const unsupported = 'Ta przeglądarka nie obsługuje WebRTC'
export const admin_loggedin = 'Jesteś zalogowany jako administrator'
export const you = 'Ty'
export const somebody = 'Ktoś'
export const send_a_message = 'Wyślij wiadomość'

export const side = {
  chat: 'Chat',
  files: 'Pliki',
  settings: 'Ustawienia',
  users: 'Użytkownicy',
}

export const connect = {
  login_title: 'Proszę się zalogować',
  invitation_title: 'Zostałeś zaproszony do tego pokoju',
  displayname: 'Podaj swoją nazwę wyświetlaną',
  password: 'Hasło',
  connect: 'Połącz',
  error: 'Błąd logowania',
  empty_displayname: 'Nazwa wyświetlana nie może być pusta.',
}

export const context = {
  ignore: 'Ignoruj',
  unignore: 'Przestań ignorować',
  mute: 'Wycisz',
  unmute: 'Wyłącz wyciszenie',
  release: 'Wymuś zwolnienie kontroli',
  take: 'Wymuś przejęcie kontroli',
  give: 'Oddaj kontrolę',
  kick: 'Wyrzuć',
  ban: 'Zablokuj IP',
  confirm: {
    kick_title: 'Wyrzucić {name}?',
    kick_text: 'Czy na pewno chcesz wyrzucić {name}?',
    ban_title: 'Zablokować {name}?',
    ban_text: 'Czy chcesz zablokować {name}? Musisz zrestartować serwer, aby cofnąć tę akcję.',
    mute_title: 'Wyciszyć {name}?',
    mute_text: 'Czy na pewno chcesz wyciszyć {name}?',
    unmute_title: 'Wyłączyć wyciszenie {name}?',
    unmute_text: 'Czy chcesz wyłączyć wyciszenie {name}?',
    button_yes: 'Tak',
    button_cancel: 'Anuluj',
  },
}

export const controls = {
  release: 'Zwolnij kontrolę',
  request: 'Poproś o kontrolę',
  lock: 'Zablokuj kontrolę',
  unlock: 'Odblokuj kontrolę',
  has: 'Masz kontrolę',
  hasnot: 'Nie masz kontroli',
  mic_on: 'Włącz mikrofon',
  mic_off: 'Wyłącz mikrofon',
  mic_error: 'Błąd mikrofonu',
}

export const locks = {
  control: {
    lock: 'Zablokuj kontrolę (dla użytkowników)',
    unlock: 'Odblokuj kontrolę (dla użytkowników)',
    locked: 'Kontrola zablokowana (dla użytkowników)',
    unlocked: 'Kontrola odblokowana (dla użytkowników)',
    notif_locked: 'zablokował kontrolę dla użytkowników',
    notif_unlocked: 'odblokował kontrolę dla użytkowników',
  },
  login: {
    lock: 'Zablokuj pokój (dla użytkowników)',
    unlock: 'Odblokuj pokój (dla użytkowników)',
    locked: 'Pokój zablokowany (dla użytkowników)',
    unlocked: 'Pokój odblokowany (dla użytkowników)',
    notif_locked: 'zablokował pokój',
    notif_unlocked: 'odblokował pokój',
  },
  file_transfer: {
    lock: 'Zablokuj transfer plików (dla użytkowników)',
    unlock: 'Odblokuj transfer plików (dla użytkowników)',
    locked: 'Transfer plików zablokowany (dla użytkowników)',
    unlocked: 'Transfer plików odblokowany (dla użytkowników)',
    notif_locked: 'zablokował transfer plików',
    notif_unlocked: 'odblokował transfer plików',
  },
}

export const setting = {
  scroll: 'Czułość przewijania',
  scroll_invert: 'Odwróć przewijanie',
  trackpad_mode: 'Tryb gładzika',
  trackpad_mode_description: 'Przesuń kursor względnie jak gładzik',
  trackpad_mode_mobile_hint: 'Dwa palce do przewijania, dotknij aby kliknąć, przytrzymaj dla prawego przycisku',
  mobile: 'Mobilny',
  desktop_only_inactive: 'Tylko dotyk',
  autoplay: 'Autoodtwarzanie wideo',
  ignore_emotes: 'Ignoruj emotki',
  chat_sound: 'Odtwarzaj dźwięk czatu',
  keyboard_layout: 'Układ klawiatury',
  broadcast_title: 'Transmisja na żywo',
}

export const connection = {
  logged_out: 'Zostałeś wylogowany.',
  reconnecting: 'Ponowne łączenie...',
  connected: 'Połączono',
  disconnected: 'Rozłączono',
  kicked: 'Zostałeś usunięty z pokoju.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} dołączył',
  disconnected: '{name} opuścił',
  controls_taken: '{name} przejął kontrolę',
  controls_taken_force: 'przejął kontrolę siłą',
  controls_taken_steal: 'przejął kontrolę od {name}',
  controls_released: '{name} zwolnił kontrolę',
  controls_released_force: 'zwolnił kontrolę siłą',
  controls_released_steal: 'zwolnił kontrolę od {name}',
  controls_given: 'oddał kontrolę {name}',
  controls_has: '{name} ma kontrolę',
  controls_has_alt: 'Ale dałem znać tej osobie, że tego chciałeś',
  controls_requesting: '{name} prosi o kontrolę',
  resolution: 'zmienił rozdzielczość na {width}x{height}@{rate}',
  banned: 'zablokował {name}',
  kicked: 'wyrzucił {name}',
  muted: 'wyciszył {name}',
  unmuted: 'wyłączył wyciszenie {name}',
}

export const files = {
  downloads: 'Pobrane',
  uploads: 'Przesłane',
  upload_here: 'Kliknij lub przeciągnij pliki tutaj, aby przesłać',
}
