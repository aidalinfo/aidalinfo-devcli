export namespace backend {
	
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

}

