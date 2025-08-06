import { TestMySQLConnection } from '../../wailsjs/go/main/App';

export interface MySQLServer {
  id: string;
  name: string;
  host: string;
  port: string;
  user: string;
  password: string;
  authDatabase?: string; // Base d'authentification (par défaut: mysql)
  isDefault?: boolean;
  createdAt: string;
  updatedAt: string;
}

const STORAGE_KEY = 'mysqlServers';
const OLD_STORAGE_KEY = 'mysql';

export class MySQLServersManager {
  /**
   * Récupère tous les serveurs MySQL stockés
   */
  static getServers(): MySQLServer[] {
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
  static getServer(id: string): MySQLServer | undefined {
    const servers = this.getServers();
    return servers.find(s => s.id === id);
  }

  /**
   * Récupère un serveur par son ID (alias pour getServer)
   */
  static getServerById(id: string): MySQLServer | undefined {
    return this.getServer(id);
  }

  /**
   * Récupère le serveur par défaut
   */
  static getDefaultServer(): MySQLServer | undefined {
    const servers = this.getServers();
    return servers.find(s => s.isDefault) || servers[0];
  }

  /**
   * Ajoute un nouveau serveur
   */
  static addServer(server: Omit<MySQLServer, 'id' | 'createdAt' | 'updatedAt'>): MySQLServer {
    const servers = this.getServers();
    
    // Si c'est le premier serveur ou si défini comme défaut, définir comme défaut
    if (servers.length === 0 || server.isDefault) {
      // Retirer le défaut des autres serveurs
      servers.forEach(s => s.isDefault = false);
    }

    const newServer: MySQLServer = {
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
  static updateServer(id: string, updates: Partial<Omit<MySQLServer, 'id' | 'createdAt'>>): MySQLServer | null {
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
   * Teste la connexion à un serveur MySQL
   */
  static async testConnection(server: Partial<MySQLServer>): Promise<{ success: boolean; message: string }> {
    try {
      await TestMySQLConnection(
        server.host || 'localhost',
        server.port || '3306',
        server.user || '',
        server.password || ''
      );
      return {
        success: true,
        message: 'Connexion réussie'
      };
    } catch (error) {
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Échec de la connexion'
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
      const imported = JSON.parse(jsonString) as MySQLServer[];
      
      if (!Array.isArray(imported)) {
        throw new Error('Format invalide: un tableau de serveurs est attendu');
      }

      // Valider la structure
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
        message: error instanceof Error ? error.message : 'Échec de l\'import'
      };
    }
  }

  /**
   * Migration des anciennes données MySQL
   */
  private static migrateOldData(): void {
    // Check for old format data in individual keys
    const mysqlHost = localStorage.getItem('mysql_host');
    const mysqlPort = localStorage.getItem('mysql_port');
    const mysqlUser = localStorage.getItem('mysql_user');
    const mysqlPassword = localStorage.getItem('mysql_password');
    
    if (mysqlHost || mysqlPort) {
      // Créer un serveur à partir des anciennes données
      const migratedServer: MySQLServer = {
        id: this.generateId(),
        name: 'Serveur MySQL par défaut',
        host: mysqlHost || 'localhost',
        port: mysqlPort || '3306',
        user: mysqlUser || '',
        password: mysqlPassword || '',
        authDatabase: 'mysql',
        isDefault: true,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      };

      this.saveServers([migratedServer]);
      
      console.log('Paramètres MySQL migrés avec succès');
    } else {
      // Pas de données à migrer
      this.saveServers([]);
    }
  }

  /**
   * Sauvegarde les serveurs dans le localStorage
   */
  private static saveServers(servers: MySQLServer[]): void {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(servers));
  }

  /**
   * Génère un ID unique
   */
  private static generateId(): string {
    return `mysql_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }
}

// Fonction utilitaire pour récupérer les paramètres de connexion MySQL formatés
export function getMySQLConnectionParams(server: MySQLServer): {
  mysqlHost: string;
  mysqlPort: string;
  mysqlUser: string;
  mysqlPassword: string;
} {
  return {
    mysqlHost: server.host,
    mysqlPort: server.port,
    mysqlUser: server.user,
    mysqlPassword: server.password
  };
}