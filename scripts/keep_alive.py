import requests
import time
import os
import json
from datetime import datetime
from urllib3.exceptions import InsecureRequestWarning

# Desabilitar avisos de SSL
requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

def keep_alive():
    # URL da API (pode ser configurada via variável de ambiente)
    api_url = os.getenv('API_URL', 'https://lta-results-api.onrender.com')
    health_check_url = f"{api_url}/health"
    
    # Intervalo entre requisições (em segundos)
    interval = 30  # Faz uma requisição a cada 30 segundos
    
    # Headers para simular um navegador
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36',
        'Accept': 'application/json',
        'Accept-Language': 'pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7',
        'Connection': 'keep-alive',
    }
    
    print(f"Iniciando keep-alive para {health_check_url}")
    print(f"Intervalo entre requisições: {interval} segundos")
    print(f"Headers: {json.dumps(headers, indent=2)}")
    
    while True:
        try:
            # Primeiro, fazer uma requisição OPTIONS
            options_response = requests.options(
                health_check_url,
                headers=headers,
                verify=False,
                timeout=10
            )
            
            # Depois, fazer a requisição GET
            response = requests.get(
                health_check_url,
                headers=headers,
                verify=False,
                timeout=10
            )
            
            current_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            
            if response.status_code == 200:
                print(f"[{current_time}] Health check bem-sucedido:")
                print(f"Status: {response.status_code}")
                print(f"Headers: {json.dumps(dict(response.headers), indent=2)}")
                print(f"Response: {json.dumps(response.json(), indent=2)}")
            else:
                print(f"[{current_time}] Health check falhou:")
                print(f"Status: {response.status_code}")
                print(f"Headers: {json.dumps(dict(response.headers), indent=2)}")
                print(f"Response: {response.text}")
                
        except Exception as e:
            print(f"[{current_time}] Erro ao fazer health check: {str(e)}")
            
        time.sleep(interval)

if __name__ == "__main__":
    keep_alive() 