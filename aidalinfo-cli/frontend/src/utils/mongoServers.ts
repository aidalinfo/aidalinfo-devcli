export interface MongoServer {
  id: string;
  name: string;
  host: string;
  port: string;
  user: string;
  password: string;
  isDefault?: boolean;
  createdAt: string;
  updatedAt: string;
}

const STORAGE_KEY = 'mongoServers';
const OLD_STORAGE_KEY = 'mongodb';

export class MongoServersManager {
  /**
   * Récupère tous les serveurs MongoDB stockés
   */
  static getServers(): MongoServer[] {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (!stored) {
      // Migration des anciennes données
      this.migrateOldData();
      return this.getServers();
    }
    try {
      return JSON.parse(stored);
    } catch {
      return [];
    }
  }

  /**
   * Récupère un serveur par son ID
   */
  static getServer(id: string): MongoServer | undefined {
    const servers = this.getServers();
    return servers.find(s => s.id === id);
  }

  /**
   * Récupère le serveur par défaut
   */
  static getDefaultServer(): MongoServer | undefined {
    const servers = this.getServers();
    return servers.find(s => s.isDefault) || servers[0];
  }

  /**
   * Ajoute un nouveau serveur
   */
  static addServer(server: Omit<MongoServer, 'id' | 'createdAt' | 'updatedAt'>): MongoServer {
    const servers = this.getServers();
    
    // Si c'est le premier serveur ou si défini comme défaut, définir comme défaut
    if (servers.length === 0 || server.isDefault) {
      // Retirer le défaut des autres serveurs
      servers.forEach(s => s.isDefault = false);
    }

    const newServer: MongoServer = {
      ...server,
      id: this.generateId(),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };

    servers.push(newServer);
    this.saveServers(servers);
    return newServer;
  }

  /**
   * Met à jour un serveur existant
   */
  static updateServer(id: string, updates: Partial<Omit<MongoServer, 'id' | 'createdAt'>>): MongoServer | null {
    const servers = this.getServers();
    const index = servers.findIndex(s => s.id === id);
    
    if (index === -1) return null;

    // Si on définit comme défaut, retirer le défaut des autres
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

  /**
   * Supprime un serveur
   */
  static deleteServer(id: string): boolean {
    const servers = this.getServers();
    const filtered = servers.filter(s => s.id !== id);
    
    if (filtered.length === servers.length) return false;

    // Si on supprime le serveur par défaut, définir le premier comme défaut
    const deletedWasDefault = servers.find(s => s.id === id)?.isDefault;
    if (deletedWasDefault && filtered.length > 0) {
      filtered[0].isDefault = true;
    }

    this.saveServers(filtered);
    return true;
  }

  /**
   * Définit un serveur comme défaut
   */
  static setDefaultServer(id: string): boolean {
    const servers = this.getServers();
    const server = servers.find(s => s.id === id);
    
    if (!server) return false;

    servers.forEach(s => s.isDefault = s.id === id);
    this.saveServers(servers);
    return true;
  }

  /**
   * Teste la connexion à un serveur MongoDB
   */
  static async testConnection(server: Partial<MongoServer>): Promise<{ success: boolean; message: string }> {
    try {
      // Ici on pourrait implémenter un appel API pour tester la connexion
      // Pour l'instant, on retourne un succès simulé
      return {
        success: true,
        message: 'Connection test not implemented yet'
      };
    } catch (error) {
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Connection failed'
      };
    }
  }

  /**
   * Exporte la configuration des serveurs
   */
  static exportServers(): string {
    const servers = this.getServers();
    return JSON.stringify(servers, null, 2);
  }

  /**
   * Importe une configuration de serveurs
   */
  static importServers(jsonString: string, replace: boolean = false): { success: boolean; message: string } {
    try {
      const imported = JSON.parse(jsonString) as MongoServer[];
      
      if (!Array.isArray(imported)) {
        throw new Error('Invalid format: expected an array of servers');
      }

      // Valider la structure
      for (const server of imported) {
        if (!server.id || !server.name || !server.host || !server.port) {
          throw new Error('Invalid server structure');
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

      return { success: true, message: 'Servers imported successfully' };
    } catch (error) {
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Import failed'
      };
    }
  }

  /**
   * Migration des anciennes données MongoDB
   */
  private static migrateOldData(): void {
    // Check for old format data in individual keys
    const mongoHost = localStorage.getItem('mongo_host');
    const mongoPort = localStorage.getItem('mongo_port');
    const mongoUser = localStorage.getItem('mongo_user');
    const mongoPassword = localStorage.getItem('mongo_password');
    
    if (mongoHost || mongoPort) {
      // Créer un serveur à partir des anciennes données
      const migratedServer: MongoServer = {
        id: this.generateId(),
        name: 'Default MongoDB Server',
        host: mongoHost || 'localhost',
        port: mongoPort || '27017',
        user: mongoUser || '',
        password: mongoPassword || '',
        isDefault: true,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      };

      this.saveServers([migratedServer]);
      
      console.log('MongoDB settings migrated successfully');
    } else {
      // Pas de données à migrer
      this.saveServers([]);
    }
  }

  /**
   * Sauvegarde les serveurs dans le localStorage
   */
  private static saveServers(servers: MongoServer[]): void {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(servers));
  }

  /**
   * Génère un ID unique
   */
  private static generateId(): string {
    return `mongo_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }
}

// Fonction utilitaire pour récupérer les paramètres de connexion MongoDB formatés
export function getMongoConnectionParams(server: MongoServer): {
  mongoHost: string;
  mongoPort: string;
  mongoUser: string;
  mongoPassword: string;
} {
  return {
    mongoHost: server.host,
    mongoPort: server.port,
    mongoUser: server.user,
    mongoPassword: server.password
  };
}