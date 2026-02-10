export interface S3Server {
  id: string;
  name: string;
  host: string;
  port: string;
  accessKey: string;
  secretKey: string;
  region: string;
  useHttps: boolean;
  bucket?: string;
  isDefault?: boolean;
  createdAt: string;
  updatedAt: string;
}

const STORAGE_KEY = 's3Servers';
const BACKUP_REPOSITORY_SERVER_KEY = 'backupRepositoryServerId';

export class S3ServersManager {
  static getServers(): S3Server[] {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (!stored) {
      this.migrateOldData();
      return this.getServers();
    }
    try {
      return JSON.parse(stored);
    } catch {
      return [];
    }
  }

  static getServer(id: string): S3Server | undefined {
    const servers = this.getServers();
    return servers.find(s => s.id === id);
  }

  static getServerById(id: string): S3Server | undefined {
    return this.getServer(id);
  }

  static getDefaultServer(): S3Server | undefined {
    const servers = this.getServers();
    return servers.find(s => s.isDefault) || servers[0];
  }

  static addServer(server: Omit<S3Server, 'id' | 'createdAt' | 'updatedAt'>): S3Server {
    const servers = this.getServers();
    
    if (servers.length === 0 || server.isDefault) {
      servers.forEach(s => s.isDefault = false);
    }

    const newServer: S3Server = {
      ...server,
      id: this.generateId(),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };

    servers.push(newServer);
    this.saveServers(servers);
    return newServer;
  }

  static updateServer(id: string, updates: Partial<Omit<S3Server, 'id' | 'createdAt'>>): S3Server | null {
    const servers = this.getServers();
    const index = servers.findIndex(s => s.id === id);
    
    if (index === -1) return null;

    if (updates.isDefault) {
      servers.forEach(s => s.isDefault = false);
    }

    servers[index] = {
      ...servers[index],
      ...updates,
      updatedAt: new Date().toISOString()
    };

    this.saveServers(servers);
    return servers[index];
  }

  static deleteServer(id: string): boolean {
    const servers = this.getServers();
    const filtered = servers.filter(s => s.id !== id);
    
    if (filtered.length === servers.length) return false;

    const deletedWasDefault = servers.find(s => s.id === id)?.isDefault;
    if (deletedWasDefault && filtered.length > 0) {
      filtered[0].isDefault = true;
    }

    this.saveServers(filtered);
    return true;
  }

  static setDefaultServer(id: string): boolean {
    const servers = this.getServers();
    const server = servers.find(s => s.id === id);
    
    if (!server) return false;

    servers.forEach(s => s.isDefault = s.id === id);
    this.saveServers(servers);
    return true;
  }

  static async testConnection(server: Partial<S3Server>): Promise<{ success: boolean; message: string }> {
    try {
      // TODO: Implémenter un appel API pour tester la connexion S3
      return {
        success: true,
        message: 'Test de connexion S3 à implémenter'
      };
    } catch (error) {
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Connexion échouée'
      };
    }
  }

  static exportServers(): string {
    const servers = this.getServers();
    return JSON.stringify(servers, null, 2);
  }

  static importServers(jsonString: string, replace: boolean = false): { success: boolean; message: string } {
    try {
      const imported = JSON.parse(jsonString) as S3Server[];
      
      if (!Array.isArray(imported)) {
        throw new Error('Format invalide: tableau de serveurs attendu');
      }

      for (const server of imported) {
        if (!server.id || !server.name || !server.host || !server.port) {
          throw new Error('Structure de serveur invalide');
        }
      }

      if (replace) {
        this.saveServers(imported);
      } else {
        const existing = this.getServers();
        const merged = [...existing];
        
        for (const server of imported) {
          if (!existing.find(s => s.id === server.id)) {
            merged.push(server);
          }
        }
        
        this.saveServers(merged);
      }

      return { success: true, message: 'Serveurs importés avec succès' };
    } catch (error) {
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Échec de l\'importation'
      };
    }
  }

  private static migrateOldData(): void {
    // Migration depuis les anciennes clés localStorage
    const s3AccessKey = localStorage.getItem('s3_access_key');
    const s3SecretKey = localStorage.getItem('s3_secret_key');
    const s3LocalAccessKey = localStorage.getItem('s3local_access_key');
    const s3LocalSecretKey = localStorage.getItem('s3local_secret_key');
    const s3Host = localStorage.getItem('s3_host');
    const s3Port = localStorage.getItem('s3_port');
    const s3Region = localStorage.getItem('s3_region');
    const s3UseHttps = localStorage.getItem('s3_use_https');
    
    const servers: S3Server[] = [];
    
    // Migrer le serveur S3 distant (Scaleway)
    if (s3AccessKey && s3SecretKey) {
      servers.push({
        id: this.generateId(),
        name: 'Scaleway S3',
        host: 's3.fr-par.scw.cloud',
        port: '443',
        accessKey: s3AccessKey,
        secretKey: s3SecretKey,
        region: 'fr-par',
        useHttps: true,
        isDefault: true,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      });
    }
    
    // Migrer le serveur S3 local (MinIO)
    if (s3LocalAccessKey && s3LocalSecretKey) {
      servers.push({
        id: this.generateId(),
        name: 'MinIO Local',
        host: s3Host || 'localhost',
        port: s3Port || '9000',
        accessKey: s3LocalAccessKey,
        secretKey: s3LocalSecretKey,
        region: s3Region || 'fr-par',
        useHttps: s3UseHttps === 'true',
        isDefault: servers.length === 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      });
    }
    
    if (servers.length > 0) {
      this.saveServers(servers);
      console.log('Configuration S3 migrée avec succès');
    } else {
      this.saveServers([]);
    }
  }

  private static saveServers(servers: S3Server[]): void {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(servers));
  }

  private static generateId(): string {
    return `s3_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }
}

export function getBackupRepositoryServerId(): string | null {
  return localStorage.getItem(BACKUP_REPOSITORY_SERVER_KEY);
}

export function setBackupRepositoryServerId(id: string): void {
  localStorage.setItem(BACKUP_REPOSITORY_SERVER_KEY, id);
}

export function clearBackupRepositoryServerId(): void {
  localStorage.removeItem(BACKUP_REPOSITORY_SERVER_KEY);
}

export function getBackupRepositoryServer(): S3Server | undefined {
  const id = getBackupRepositoryServerId();
  if (!id) return undefined;
  const server = S3ServersManager.getServer(id);
  if (!server) {
    clearBackupRepositoryServerId();
  }
  return server;
}

export function getS3ConnectionParams(server: S3Server): {
  host: string;
  port: string;
  accessKey: string;
  secretKey: string;
  region: string;
  useHttps: boolean;
  bucket?: string;
} {
  return {
    host: server.host,
    port: server.port,
    accessKey: server.accessKey,
    secretKey: server.secretKey,
    region: server.region,
    useHttps: server.useHttps,
    bucket: server.bucket
  };
}
