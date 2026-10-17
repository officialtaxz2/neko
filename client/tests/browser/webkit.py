# Opens URL (argv[1]) for SECONDS (argv[2]) in WebKitGTK with MSE enabled.
# WebKitGTK plays H.264/AAC through GStreamer, unlike Playwright's Chromium.
import sys
import gi
gi.require_version('WebKit2', '4.1')
gi.require_version('Gtk', '3.0')
from gi.repository import GLib, Gtk, WebKit2

context = WebKit2.WebContext.get_default()
context.set_tls_errors_policy(WebKit2.TLSErrorsPolicy.IGNORE)  # httptest certificate
view = WebKit2.WebView.new_with_context(context)
settings = view.get_settings()
settings.set_enable_mediasource(True)
settings.set_media_playback_requires_user_gesture(False)
window = Gtk.Window()
window.add(view)
window.set_default_size(1280, 720)
window.show_all()
view.load_uri(sys.argv[1])
GLib.timeout_add_seconds(int(sys.argv[2]), Gtk.main_quit)
Gtk.main()
