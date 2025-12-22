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
API_URL = "http://localhost:8080"
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

# --- Server Manager (sin cambios mayores) ---
class ServerManager:
    def __init__(self):
        self.process = None
        self.running = False
        self.logs = []
        self.log_lock = threading.Lock()

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
                r = requests.get(f"{API_URL}/health", timeout=1)
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

# --- API Client Expandido ---
class ApiClient:
    def __init__(self):
        self.base = API_URL
        self.headers = {"Content-Type": "application/json", "X-Api-Key": "kero-kero-api-key"}
        self._load_env_key()

    def _load_env_key(self):
        env_path = os.path.join(SERVER_DIR, ".env")
        if os.path.exists(env_path):
            with open(env_path, 'r') as f:
                for line in f:
                    if line.startswith("API_KEY="):
                        self.headers["X-Api-Key"] = line.split("=")[1].strip()

    def _req(self, method, endpoint, data=None):
        url = f"{self.base}{endpoint}"
        try:
            if method == 'GET': r = requests.get(url, headers=self.headers)
            elif method == 'POST': r = requests.post(url, json=data, headers=self.headers)
            elif method == 'PUT': r = requests.put(url, json=data, headers=self.headers)
            elif method == 'DELETE': r = requests.delete(url, headers=self.headers)
            
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
    
    def react(self, iid, msg_id, emoji):
        return self._req('POST', f'/instances/{iid}/messages/react', {"message_id": msg_id, "reaction": emoji})

    def revoke(self, iid, msg_id):
        return self._req('POST', f'/instances/{iid}/messages/revoke', {"message_id": msg_id})

    # -- Contacts --
    def check_contacts(self, iid, phones):
        return self._req('POST', f'/instances/{iid}/contacts/check', {"phones": phones})
    
    def get_contacts(self, iid): return self._req('GET', f'/instances/{iid}/contacts')
    def get_contact_info(self, iid, phone): return self._req('GET', f'/instances/{iid}/contacts/{phone}')
    def get_profile_pic(self, iid, phone): return self._req('GET', f'/instances/{iid}/contacts/{phone}/profile-picture')
    def block_contact(self, iid, phone): return self._req('POST', f'/instances/{iid}/contacts/block', {"phone": phone})
    def unblock_contact(self, iid, phone): return self._req('POST', f'/instances/{iid}/contacts/unblock', {"phone": phone})

    # -- Groups --
    def list_groups(self, iid): return self._req('GET', f'/instances/{iid}/groups')
    def create_group(self, iid, name, participants):
        return self._req('POST', f'/instances/{iid}/groups', {"subject": name, "participants": participants})
    def get_group_info(self, iid, gid): return self._req('GET', f'/instances/{iid}/groups/{gid}')
    def get_group_invite(self, iid, gid): return self._req('GET', f'/instances/{iid}/groups/{gid}/invite')
    def leave_group(self, iid, gid): return self._req('POST', f'/instances/{iid}/groups/{gid}/leave')

# --- UI Helpers ---
def clear(): os.system('cls' if os.name == 'nt' else 'clear')
def prompt(p): return input(f"{Colors.GREEN}?{Colors.ENDC} {p}: ").strip()
def pause(): input("\nPresiona ENTER para continuar...")

def pretty_print(data):
    try:
        if isinstance(data, str): data = json.loads(data)
        print(f"{Colors.CYAN}{json.dumps(data, indent=2, ensure_ascii=False)}{Colors.ENDC}")
    except:
        print(data)

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
        print("0. Volver")
        op = prompt("Opción")
        
        if op == '0': return
        if op == '1': pretty_print(api.list_instances())
        if op == '2': pretty_print(api.create_instance(prompt("ID Nueva Instancia")))
        if op == '3': pretty_print(api.delete_instance(prompt("ID a Eliminar")))
        
        if op == '4':
            iid = prompt("ID Instancia")
            print_info("Conectando...")
            print(api.connect_instance(iid))
            print_info("Obteniendo QR...")
            r = api.get_qr(iid)
            if 'qr_code' in r: show_qr_img(r['qr_code'])
            else: pretty_print(r)
            
        if op == '5': pretty_print(api.disconnect_instance(prompt("ID Instancia")))
        if op == '6': pretty_print(api.get_status(prompt("ID Instancia")))
        if op not in ['1','2','3','4','5','6']: pause()
        else: pause()

def module_messages(api, iid):
    if not iid: return print_error("Selecciona una instancia primero!")
    
    while True:
        print_header(f"MENSAJERÍA (Instancia: {iid})")
        print("1. Texto Plano")
        print("2. Imagen (URL)")
        print("3. Video (URL)")
        print("4. Audio (URL)")
        print("5. Documento (URL)")
        print("6. Ubicación")
        print("7. Contacto (VCard)")
        print("8. Reaccionar")
        print("9. Revocar (Eliminar)")
        print("0. Volver")
        op = prompt("Opción")
        
        if op == '0': return
        
        to = prompt("Destinatario (Tel: 521...)")
        
        if op == '1': pretty_print(api.send_text(iid, to, prompt("Mensaje")))
        if op == '2': pretty_print(api.send_image(iid, to, prompt("URL Imagen"), prompt("Caption")))
        if op == '3': pretty_print(api.send_video(iid, to, prompt("URL Video"), prompt("Caption")))
        if op == '4': pretty_print(api.send_audio(iid, to, prompt("URL Audio")))
        if op == '5': pretty_print(api.send_document(iid, to, prompt("URL Doc"), prompt("Filename")))
        if op == '6': pretty_print(api.send_location(iid, to, float(prompt("Lat")), float(prompt("Long")), prompt("Nombre"), prompt("Dirección")))
        if op == '7': pretty_print(api.send_contact(iid, to, prompt("VCard Data")))
        if op == '8': pretty_print(api.react(iid, prompt("Message ID"), prompt("Emoji")))
        if op == '9': pretty_print(api.revoke(iid, prompt("Message ID")))
        
        pause()

def module_contacts_groups(api, iid):
    if not iid: return print_error("Selecciona una instancia primero!")
    
    while True:
        print_header(f"CONTACTOS Y GRUPOS ({iid})")
        print("1. Listar Contactos")
        print("2. Info de un Contacto")
        print("3. Verificar Números")
        print("4. Obtener Foto Perfil")
        print("5. Bloquear Contacto")
        print("6. Desbloquear Contacto")
        print("---")
        print("7. Listar Grupos")
        print("8. Info Grupo")
        print("9. Link Invitación Grupo")
        print("10. Crear Grupo")
        print("0. Volver")
        op = prompt("Opción")
        
        if op == '0': return
        
        if op == '1': pretty_print(api.get_contacts(iid))
        if op == '2': pretty_print(api.get_contact_info(iid, prompt("Teléfono")))
        if op == '3': pretty_print(api.check_contacts(iid, prompt("Números (sep por coma)").split(',')))
        if op == '4': pretty_print(api.get_profile_pic(iid, prompt("Teléfono")))
        if op == '5': pretty_print(api.block_contact(iid, prompt("Teléfono")))
        if op == '6': pretty_print(api.unblock_contact(iid, prompt("Teléfono")))
        
        if op == '7': pretty_print(api.list_groups(iid))
        if op == '8': pretty_print(api.get_group_info(iid, prompt("Group ID")))
        if op == '9': pretty_print(api.get_group_invite(iid, prompt("Group ID")))
        if op == '10': pretty_print(api.create_group(iid, prompt("Nombre"), prompt("Participantes (sep por coma)").split(',')))
        
        pause()

# --- Main Loop ---
def main():
    server = ServerManager()
    if not server.start(): return
    
    api = ApiClient()
    current_iid = None
    
    # Auto-select single instance
    try:
        insts = api.list_instances()
        if len(insts) == 1: current_iid = insts[0]['instance_id']
    except: pass
    
    while True:
        clear()
        print_header("KERO-KERO FULL CLI TESTER")
        print(f"Instancia Activa: {Colors.CYAN}{current_iid or 'Ninguna'}{Colors.ENDC}")
        print("\nmenú Principal:")
        print("1. Instancias (Conexión/QR)")
        print("2. Mensajería (Texto, Media, Loc)")
        print("3. Contactos y Grupos")
        print("4. Cambiar Instancia Activa")
        print("5. Ver Logs Servidor")
        print("0. Salir")
        
        op = prompt("Opción")
        
        if op == '1': module_instances(api)
        elif op == '2': module_messages(api, current_iid)
        elif op == '3': module_contacts_groups(api, current_iid)
        elif op == '4':
            l = api.list_instances()
            print("Disponibles:", [i['instance_id'] for i in l])
            current_iid = prompt("ID Instancia")
        elif op == '5':
            print("\n".join(server.get_logs()[-30:]))
            pause()
        elif op == '0': break
        
    server.stop()

if __name__ == "__main__":
    try: main()
    except KeyboardInterrupt: sys.exit(0)
