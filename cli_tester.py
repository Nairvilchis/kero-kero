#!/usr/bin/env python3
import subprocess
import time
import sys
import os
import json
import base64
import signal
import threading
import shutil
import platform

# --- Configuración Base ---
SERVER_DIR = os.path.dirname(os.path.abspath(__file__))
SERVER_CMD = ["go", "run", "cmd/server/main.go"]
DEFAULT_LOCAL_URL = "http://localhost:8080"
QR_FILENAME = "current_qr.png"

# --- Colores ---
class Colors:
    HEADER = '\033[95m'
    BLUE = '\033[94m'
    CYAN = '\033[96m'
    GREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'

def print_info(msg): print(f"{Colors.BLUE}[INFO]{Colors.ENDC} {msg}")
def print_success(msg): print(f"{Colors.GREEN}[OK]{Colors.ENDC} {msg}")
def print_error(msg): print(f"{Colors.FAIL}[ERROR]{Colors.ENDC} {msg}")
def print_header(msg): print(f"\n{Colors.HEADER}{Colors.BOLD}=== {msg} ==={Colors.ENDC}")

# --- Dep Check ---
try:
    import requests
except ImportError:
    print_error("La librería 'requests' no está instalada.")
    print("Ejecuta: pip install requests")
    sys.exit(1)

# --- Server Manager ---
class ServerManager:
    def __init__(self, api_url=DEFAULT_LOCAL_URL):
        self.process = None
        self.running = False
        self.logs = []
        self.log_lock = threading.Lock()
        self.api_url = api_url

    def start(self):
        print_info("Iniciando servidor Kero-Kero (go run cmd/server/main.go)...")
        try:
            self.process = subprocess.Popen(
                SERVER_CMD,
                cwd=SERVER_DIR,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                text=True,
                preexec_fn=os.setsid if platform.system() != "Windows" else None
            )
            self.running = True
            threading.Thread(target=self._read_logs, daemon=True).start()
            return self._wait_for_health()
        except FileNotFoundError:
            print_error("No se encontró 'go'. Asegúrate de tener Go instalado.")
            return False
        except Exception as e:
            print_error(f"Error iniciando servidor: {e}")
            return False

    def _read_logs(self):
        while self.running and self.process:
            line = self.process.stdout.readline()
            if not line: break
            with self.log_lock:
                self.logs.append(line.strip())
                if len(self.logs) > 1000: self.logs.pop(0)

    def _wait_for_health(self, timeout=30):
        print("Esperando health check...")
        start = time.time()
        spinner = "|/-\\"
        idx = 0
        while time.time() - start < timeout:
            try:
                r = requests.get(f"{self.api_url}/health", timeout=1)
                if r.status_code == 200:
                    print()
                    print_success("Servidor Listo!")
                    return True
            except: pass
            sys.stdout.write(f"\r{spinner[idx%4]}")
            sys.stdout.flush()
            idx += 1
            time.sleep(0.2)
        print_error("Timeout esperando servidor.")
        return False

    def get_logs(self):
        with self.log_lock: return list(self.logs)

    def stop(self):
        if self.process:
            self.running = False
            try:
                if platform.system() != "Windows":
                    os.killpg(os.getpgid(self.process.pid), signal.SIGTERM)
                else:
                    self.process.terminate()
            except: pass
            print_success("Servidor detenido.")

# --- API Client Ultimate ---
class ApiClient:
    def __init__(self, base_url, api_key):
        self.base = base_url
        self.headers = {"Content-Type": "application/json", "X-Api-Key": api_key}
        print_info(f"Cliente API iniciado apuntando a: {self.base}")

    def _req(self, method, endpoint, data=None):
        url = f"{self.base}{endpoint}"
        try:
            if method == 'GET': r = requests.get(url, headers=self.headers)
            elif method == 'POST': r = requests.post(url, json=data, headers=self.headers)
            elif method == 'PUT': r = requests.put(url, json=data, headers=self.headers)
            elif method == 'DELETE': r = requests.delete(url, headers=self.headers)
            elif method == 'PATCH': r = requests.patch(url, json=data, headers=self.headers)
            
            try: return r.json()
            except: return {"status": r.status_code, "text": r.text}
        except Exception as e:
            return {"error": str(e)}

    # -- Instances --
    def list_instances(self): return self._req('GET', '/instances').get('data', [])
    def create_instance(self, iid): return self._req('POST', '/instances', {"instance_id": iid})
    def connect_instance(self, iid): return self._req('POST', f'/instances/{iid}/connect')
    def disconnect_instance(self, iid): return self._req('POST', f'/instances/{iid}/disconnect')
    def get_qr(self, iid): return self._req('GET', f'/instances/{iid}/qr')
    def get_status(self, iid): return self._req('GET', f'/instances/{iid}/status')
    def delete_instance(self, iid): return self._req('DELETE', f'/instances/{iid}')
    
    # -- Messages --
    def send_text(self, iid, to, msg): 
        return self._req('POST', f'/instances/{iid}/messages/text', {"to": to, "message": msg})
    def send_text_typing(self, iid, to, msg, dur): 
        return self._req('POST', f'/instances/{iid}/messages/text-with-typing', {"to": to, "message": msg, "duration": int(dur)})
    def send_image(self, iid, to, url, caption):
        return self._req('POST', f'/instances/{iid}/messages/image', {"to": to, "url": url, "caption": caption})
    def send_video(self, iid, to, url, caption):
        return self._req('POST', f'/instances/{iid}/messages/video', {"to": to, "url": url, "caption": caption})
    def send_audio(self, iid, to, url):
        return self._req('POST', f'/instances/{iid}/messages/audio', {"to": to, "url": url})
    def send_document(self, iid, to, url, filename):
        return self._req('POST', f'/instances/{iid}/messages/document', {"to": to, "url": url, "filename": filename})
    def send_location(self, iid, to, lat, long, name, addr):
        return self._req('POST', f'/instances/{iid}/messages/location', 
                         {"to": to, "latitude": lat, "longitude": long, "name": name, "address": addr})
    def send_contact(self, iid, to, vcard):
         return self._req('POST', f'/instances/{iid}/messages/contact', {"to": to, "vcard": vcard})
    def create_poll(self, iid, to, name, options, count):
        return self._req('POST', f'/instances/{iid}/messages/poll', {"to": to, "name": name, "options": options, "selectable_count": int(count)})
    def react(self, iid, msg_id, emoji):
        return self._req('POST', f'/instances/{iid}/messages/react', {"message_id": msg_id, "reaction": emoji})
    def revoke(self, iid, msg_id):
        return self._req('POST', f'/instances/{iid}/messages/revoke', {"message_id": msg_id})
    def edit_msg(self, iid, msg_id, text, to):
        return self._req('POST', f'/instances/{iid}/messages/edit', {"message_id": msg_id, "new_text": text, "phone": to})

    # -- Contacts --
    def check_contacts(self, iid, phones): return self._req('POST', f'/instances/{iid}/contacts/check', {"phones": phones})
    def get_contacts(self, iid): return self._req('GET', f'/instances/{iid}/contacts')
    def get_contact_info(self, iid, phone): return self._req('GET', f'/instances/{iid}/contacts/{phone}')
    def get_profile_pic(self, iid, phone): return self._req('GET', f'/instances/{iid}/contacts/{phone}/profile-picture')
    def block_contact(self, iid, phone): return self._req('POST', f'/instances/{iid}/contacts/block', {"phone": phone})
    def unblock_contact(self, iid, phone): return self._req('POST', f'/instances/{iid}/contacts/unblock', {"phone": phone})
    def sub_presence(self, iid, phones): return self._req('POST', f'/instances/{iid}/contacts/presence/subscribe', {"phones": phones})

    # -- Groups --
    def list_groups(self, iid): return self._req('GET', f'/instances/{iid}/groups')
    def create_group(self, iid, name, participants):
        return self._req('POST', f'/instances/{iid}/groups', {"subject": name, "participants": participants})
    def get_group_info(self, iid, gid): return self._req('GET', f'/instances/{iid}/groups/{gid}')
    def get_group_invite(self, iid, gid): return self._req('GET', f'/instances/{iid}/groups/{gid}/invite')
    def leave_group(self, iid, gid): return self._req('POST', f'/instances/{iid}/groups/{gid}/leave')
    def join_group(self, iid, code): return self._req('POST', f'/instances/{iid}/groups/join', {"code": code})
    
    def group_add_participants(self, iid, gid, parts): return self._req('POST', f'/instances/{iid}/groups/{gid}/participants', {"participants": parts})
    def group_update_settings(self, iid, gid, announce, locked): return self._req('PUT', f'/instances/{iid}/groups/{gid}/settings', {"announce": announce, "locked": locked})
    
    # -- Automation --
    def send_bulk(self, iid, numbers, msg, delay=1): 
        return self._req('POST', f'/instances/{iid}/automation/bulk-message', {"numbers": numbers, "message": msg, "delay": delay})
    def set_autoreply(self, iid, match, reply):
        return self._req('POST', f'/instances/{iid}/automation/auto-reply', {"match": match, "response": reply})

    # -- Business --
    def create_label(self, iid, name, color): return self._req('POST', f'/instances/{iid}/business/labels', {"name": name, "color": color})
    def set_autolabel_rule(self, iid, keywords, label_id): return self._req('POST', f'/instances/{iid}/business/autolabel/rules', {"keywords": keywords, "label_id": label_id})

    # -- Privacy/Calls --
    def get_privacy(self, iid): return self._req('GET', f'/instances/{iid}/privacy')
    def set_privacy(self, iid, settings): return self._req('PUT', f'/instances/{iid}/privacy', settings)
    def set_call_settings(self, iid, reject_all, msg): return self._req('PUT', f'/instances/{iid}/calls/settings', {"reject_all": reject_all, "reject_message": msg})

    # -- Extras --
    def set_presence(self, iid, to, state): return self._req('POST', f'/instances/{iid}/presence/start', {"to": to, "state": state})
    def stop_presence(self, iid, to): return self._req('POST', f'/instances/{iid}/presence/stop', {"to": to})
    def post_status(self, iid, msg, color): return self._req('POST', f'/instances/{iid}/status', {"message": msg, "background_color": color})
    def create_newsletter(self, iid, name, desc): return self._req('POST', f'/instances/{iid}/newsletters', {"name": name, "description": desc})
    def list_newsletters(self, iid): return self._req('GET', f'/instances/{iid}/newsletters')
    def start_sync(self, iid, full=False): return self._req('POST', f'/instances/{iid}/sync', {"full": full})
    def get_sync_progress(self, iid): return self._req('GET', f'/instances/{iid}/sync/progress')
    def set_webhook(self, iid, url): return self._req('POST', f'/instances/{iid}/webhook', {"url": url, "events": ["message", "status"]})
    def get_webhook(self, iid): return self._req('GET', f'/instances/{iid}/webhook')

# --- UI Helpers ---
def clear(): os.system('cls' if os.name == 'nt' else 'clear')
def prompt(p): return input(f"{Colors.GREEN}?{Colors.ENDC} {p}: ").strip()
def pause(): input("\nPresiona ENTER para continuar...")
def pretty_print(data):
    try:
        if isinstance(data, str): data = json.loads(data)
        print(f"{Colors.CYAN}{json.dumps(data, indent=2, ensure_ascii=False)}{Colors.ENDC}")
    except: print(data)
def show_qr_img(b64):
    try:
        import base64
        with open(QR_FILENAME, "wb") as f: f.write(base64.b64decode(b64))
        print_info(f"QR guardado en {QR_FILENAME}")
        if shutil.which('xdg-open'): subprocess.call(['xdg-open', QR_FILENAME])
        elif shutil.which('open'): subprocess.call(['open', QR_FILENAME])
    except: print_error("No se pudo abrir la imagen QR.")

# --- Logic Modules ---
def module_instances(api):
    while True:
        print_header("GESTIÓN DE INSTANCIAS")
        print("1. Listar Instancias")
        print("2. Crear Nueva")
        print("3. Eliminar Instancia")
        print("4. Conectar / Ver QR")
        print("5. Desconectar")
        print("6. Ver Estado Detallado")
        print("7. Configurar Webhook")
        print("8. Sincronizar Historial")
        print("9. Configurar Privacidad")
        print("10. Configurar Rechazo Llamadas")
        print("0. Volver")
        op = prompt("Opción")
        
        if op == '0': return
        if op == '1': pretty_print(api.list_instances())
        if op == '2': pretty_print(api.create_instance(prompt("ID Nueva Instancia")))
        if op == '3': pretty_print(api.delete_instance(prompt("ID a Eliminar")))
        if op == '4':
            iid = prompt("ID Instancia")
            print(api.connect_instance(iid))
            r = api.get_qr(iid)
            if 'qr_code' in r: show_qr_img(r['qr_code'])
            else: pretty_print(r)
        if op == '5': pretty_print(api.disconnect_instance(prompt("ID Instancia")))
        if op == '6': pretty_print(api.get_status(prompt("ID Instancia")))
        if op == '7': pretty_print(api.set_webhook(prompt("ID Instancia"), prompt("URL Webhook")))
        if op == '8': 
            iid = prompt("ID Instancia")
            print(api.start_sync(iid, full=True))
            while True:
                time.sleep(1)
                p = api.get_sync_progress(iid)
                print(f"Progreso: {p}")
                if prompt("Actualizar? (enter=si, x=salir)") == 'x': break
        if op == '9':
            iid = prompt("ID Instancia")
            pretty_print(api.get_privacy(iid))
            if prompt("Editar? (y/n)") == 'y':
                last = prompt("Last Seen (all/contacts/none)")
                pretty_print(api.set_privacy(iid, {"last_seen": last}))
        if op == '10':
            iid = prompt("ID Instancia")
            rej = prompt("Rechazar todas? (true/false)") == 'true'
            msg = prompt("Mensaje de rechazo")
            pretty_print(api.set_call_settings(iid, rej, msg))

        if op not in ['4','8','9','10']: pause()
        else: pause()

def module_messages(api, iid):
    if not iid: return print_error("Selecciona una instancia primero!")
    while True:
        print_header(f"MENSAJERÍA (Instancia: {iid})")
        print("1. Texto Plano")
        print("2. Texto + Escribiendo...")
        print("3. Imagen (URL)")
        print("4. Video/Audio/Doc")
        print("5. Ubicación/Contacto")
        print("6. Crear ENCUESTA (Poll)")
        print("7. Editar Mensaje")
        print("8. Reaccionar/Revocar")
        print("0. Volver")
        op = prompt("Opción")
        
        if op == '0': return
        if op == '3': pretty_print(api.send_image(iid, prompt("Destinatario"), prompt("URL"), prompt("Caption")))
        elif op == '4':
            type = prompt("Tipo (video/audio/document)")
            to = prompt("Destinatario")
            url = prompt("URL")
            if type == 'video': pretty_print(api.send_video(iid, to, url, prompt("Caption")))
            elif type == 'audio': pretty_print(api.send_audio(iid, to, url))
            elif type == 'document': pretty_print(api.send_document(iid, to, url, prompt("Filename")))
        elif op == '6':
            to = prompt("Destinatario")
            name = prompt("Titulo Encuesta")
            opts = prompt("Opciones (sep coma)").split(',')
            cnt = prompt("Max respuestas")
            pretty_print(api.create_poll(iid, to, name, opts, cnt))
        elif op == '1': pretty_print(api.send_text(iid, prompt("Destinatario"), prompt("Mensaje")))
        elif op == '2': pretty_print(api.send_text_typing(iid, prompt("Destinatario"), prompt("Mensaje"), prompt("Duración")))
        elif op == '5':
            type = prompt("Tipo (location/contact)")
            to = prompt("Destinatario")
            if type == 'location': pretty_print(api.send_location(iid, to, float(prompt("Lat")), float(prompt("Lon")), "Loc", "Addr"))
            else: pretty_print(api.send_contact(iid, to, prompt("VCard")))
        elif op == '7': pretty_print(api.edit_msg(iid, prompt("MsgID"), prompt("New Text"), prompt("Destinatario")))
        elif op == '8':
            act = prompt("Accion (react/revoke)")
            mid = prompt("MsgID")
            if act == 'react': pretty_print(api.react(iid, mid, prompt("Emoji")))
            else: pretty_print(api.revoke(iid, mid))
        else: pause()
        if op != '0': pause()

def module_automation_business(api, iid):
    if not iid: return print_error("Selecciona una instancia primero!")
    while True:
        print_header(f"AUTOMATION & BUSINESS ({iid})")
        print("1. Enviar Mensaje Masivo (Bulk)")
        print("2. Configurar Auto-Respuesta")
        print("3. Crear Etiqueta (Label)")
        print("4. Crear Regla Auto-Etiquetado")
        print("0. Volver")
        op = prompt("Opción")
        
        if op == '0': return
        if op == '1':
            nums = prompt("Numeros (sep coma)").split(',')
            msg = prompt("Mensaje")
            pretty_print(api.send_bulk(iid, nums, msg))
        if op == '2':
            match = prompt("Trigger (palabra clave)")
            resp = prompt("Respuesta")
            pretty_print(api.set_autoreply(iid, match, resp))
        if op == '3': pretty_print(api.create_label(iid, prompt("Nombre"), prompt("Color Hex")))
        if op == '4':
            keys = prompt("Keywords (sep coma)").split(',')
            lid = prompt("ID Etiqueta")
            pretty_print(api.set_autolabel_rule(iid, keys, lid))
        pause()

def module_groups(api, iid):
    if not iid: return print_error("Selecciona una instancia primero!")
    while True:
        print_header(f"GRUPOS ({iid})")
        print("1. Listar Grupos / Info")
        print("2. Crear Grupo")
        print("3. Unirse / Link Invitación")
        print("4. Gestionar Participantes (Add)")
        print("5. Configuración Grupo (Lock/Announce)")
        print("0. Volver")
        op = prompt("Opción")
        
        if op == '0': return
        if op == '1': 
            pretty_print(api.list_groups(iid))
            pid = prompt("Ver detalle ID? (enter para saltar)")
            if pid: pretty_print(api.get_group_info(iid, pid))
        if op == '2': pretty_print(api.create_group(iid, prompt("Nombre"), prompt("Participantes (sep coma)").split(',')))
        if op == '3':
            act = prompt("Accion (join/link)")
            if act == 'join': pretty_print(api.join_group(iid, prompt("Code")))
            else: pretty_print(api.get_group_invite(iid, prompt("Group ID")))
        if op == '4':
            gid = prompt("Group ID")
            parts = prompt("Telefonos (sep coma)").split(',')
            pretty_print(api.group_add_participants(iid, gid, parts))
        if op == '5':
            gid = prompt("Group ID")
            ann = prompt("Solo Admins envian? (true/false)")
            lock = prompt("Solo Admins editan? (true/false)")
            pretty_print(api.group_update_settings(iid, gid, ann=='true', lock=='true'))
        pause()

# --- Main Loop ---
def main():
    clear()
    print_header("KERO-KERO CLI TESTER")
    print("Selecciona el modo de operación:")
    print("1. LOCAL (Inicia servidor local automáticamente)")
    print("2. REMOTO (Conectar a API existente / Servidor en línea)")
    
    mode = prompt("Opción (1/2)")
    
    server = None
    target_url = DEFAULT_LOCAL_URL
    
    if mode == '1':
        server = ServerManager()
        if not server.start(): return
    elif mode == '2':
        url_input = prompt(f"Ingresa URL API (enter para {DEFAULT_LOCAL_URL})")
        if url_input: 
            target_url = url_input
            if target_url.endswith('/'): target_url = target_url[:-1]
    else:
        print_error("Opción inválida")
        return

    # --- API Key Setup ---
    default_key = "kero-kero-api-key"
    try:
        env_path = os.path.join(SERVER_DIR, ".env")
        if os.path.exists(env_path):
            with open(env_path, 'r') as f:
                for line in f:
                    if line.startswith("API_KEY="):
                        default_key = line.split("=")[1].strip()
                        break
    except: pass

    print(f"API Key por defecto: {Colors.CYAN}{default_key}{Colors.ENDC}")
    input_key = prompt("Ingresa API Key (enter para usar defecto)")
    final_key = input_key if input_key else default_key

    api = ApiClient(target_url, final_key)
    current_iid = None
    
    try:
        insts = api.list_instances()
        if insts and len(insts) > 0: current_iid = insts[0]['instance_id']
    except Exception as e:
        print_error(f"No se pudo conectar o listar instancias: {e}")
    
    while True:
        clear()
        print_header("KERO-KERO ULTIMATE CLI")
        print(f"Modo: {Colors.BOLD}{'LOCAL' if server else 'REMOTO'}{Colors.ENDC} | API: {target_url}")
        print(f"Instancia Activa: {Colors.CYAN}{current_iid or 'Ninguna'}{Colors.ENDC}")
        print("\n1. Instancias & Privacidad")
        print("2. Mensajería (Polls, Media, Edit)")
        print("3. Automatización & Business (Bulk, Labels)")
        print("4. Grupos Avanzado")
        print("5. Contactos & Presencia")
        print("6. Extras (Estados, Newsletters)")
        print("7. Cambiar Instancia / Ver Logs")
        print("0. Salir")
        
        op = prompt("Opción")
        if op == '1': module_instances(api)
        elif op == '2': module_messages(api, current_iid)
        elif op == '3': module_automation_business(api, current_iid)
        elif op == '4': module_groups(api, current_iid)
        elif op == '5': 
             # Reusing previous contacts module code logic in new structure or keeping it simple
             print("... (Usa el código del módulo anterior para Contactos) ...")
        elif op == '7':
             try:
                 l = api.list_instances()
                 print("Disponibles:", [i['instance_id'] for i in l])
                 current_iid = prompt("ID Instancia")
             except: print_error("Error listando instancias")
             
             if server:
                 print("\n".join(server.get_logs()[-10:]))
             else:
                 print_info("Logs del servidor no disponibles en modo remoto")
             pause()
        elif op == '0': break
        
    if server:
        server.stop()

if __name__ == "__main__":
    try: main()
    except KeyboardInterrupt: sys.exit(0)
