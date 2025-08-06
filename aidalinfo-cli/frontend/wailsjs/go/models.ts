export namespace backend {
	
	export class BackupInfo {
	    name: string;
	    size: number;
	    lastModified: string;
	
	    static createFrom(source: any = {}) {
	        return new BackupInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.size = source["size"];
	        this.lastModified = source["lastModified"];
	    }
	}
	export class Commit {
	    Date: string;
	    Author: string;
	    Message: string;
	    Submodule: string;
	    Branch: string;
	
	    static createFrom(source: any = {}) {
	        return new Commit(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Date = source["Date"];
	        this.Author = source["Author"];
	        this.Message = source["Message"];
	        this.Submodule = source["Submodule"];
	        this.Branch = source["Branch"];
	    }
	}
	export class S3Credentials {
	    accessKey: string;
	    secretKey: string;
	
	    static createFrom(source: any = {}) {
	        return new S3Credentials(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.accessKey = source["accessKey"];
	        this.secretKey = source["secretKey"];
	    }
	}
	export class TagsResult {
	    vTags: string[];
	    rcTags: string[];
	
	    static createFrom(source: any = {}) {
	        return new TagsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vTags = source["vTags"];
	        this.rcTags = source["rcTags"];
	    }
	}
	export class UpdateInfo {
	    currentVersion: string;
	    latestVersion: string;
	    updateAvailable: boolean;
	    downloadUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.updateAvailable = source["updateAvailable"];
	        this.downloadUrl = source["downloadUrl"];
	    }
	}

}

