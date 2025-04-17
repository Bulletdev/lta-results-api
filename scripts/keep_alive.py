import requests
import time
import os
from datetime import datetime

def keep_alive():
    # URL da API (pode ser configurada via variável de ambiente)
    api_url = os.getenv('API_URL', 'https://lta-results-api.onrender.com')
    health_check_url = f"{api_url}/health"
    
    # Intervalo entre requisições (em segundos)
    interval = 30  # Faz uma requisição a cada 30 segundos
    
    print(f"Iniciando keep-alive para {health_check_url}")
    print(f"Intervalo entre requisições: {interval} segundos")
    
    while True:
        try:
            response = requests.get(health_check_url)
            current_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            
            if response.status_code == 200:
                print(f"[{current_time}] Health check bem-sucedido: {response.json()}")
            else:
                print(f"[{current_time}] Health check falhou: Status {response.status_code}")
                
        except Exception as e:
            print(f"[{current_time}] Erro ao fazer health check: {str(e)}")
            
        time.sleep(interval)

if __name__ == "__main__":
    keep_alive() 