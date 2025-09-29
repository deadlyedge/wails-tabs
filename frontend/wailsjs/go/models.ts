export namespace config {
	
	export class DatabaseConfig {
	    BaseFolder: string;
	    FileName: string;
	
	    static createFrom(source: any = {}) {
	        return new DatabaseConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BaseFolder = source["BaseFolder"];
	        this.FileName = source["FileName"];
	    }
	}
	export class HistoryConfig {
	    LastSourceFolder: string[];
	
	    static createFrom(source: any = {}) {
	        return new HistoryConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.LastSourceFolder = source["LastSourceFolder"];
	    }
	}
	export class ScanConfig {
	    SourceFolders: string[];
	    IncludeExtensions: string[];
	    FollowSymlinks: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ScanConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SourceFolders = source["SourceFolders"];
	        this.IncludeExtensions = source["IncludeExtensions"];
	        this.FollowSymlinks = source["FollowSymlinks"];
	    }
	}
	export class TargetConfig {
	    BaseFolder: string;
	    Pattern: string;
	
	    static createFrom(source: any = {}) {
	        return new TargetConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BaseFolder = source["BaseFolder"];
	        this.Pattern = source["Pattern"];
	    }
	}
	export class Settings {
	    Database: DatabaseConfig;
	    History: HistoryConfig;
	    Scan: ScanConfig;
	    Target: TargetConfig;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Database = this.convertValues(source["Database"], DatabaseConfig);
	        this.History = this.convertValues(source["History"], HistoryConfig);
	        this.Scan = this.convertValues(source["Scan"], ScanConfig);
	        this.Target = this.convertValues(source["Target"], TargetConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace media {
	
	export class MoveRequest {
	    mediaId: number;
	
	    static createFrom(source: any = {}) {
	        return new MoveRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mediaId = source["mediaId"];
	    }
	}
	export class Summary {
	    filesDiscovered: number;
	    filesPersisted: number;
	    filesSkipped: number;
	    errors: string[];
	    durationMs: number;
	    duplicateGroups: number;
	
	    static createFrom(source: any = {}) {
	        return new Summary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filesDiscovered = source["filesDiscovered"];
	        this.filesPersisted = source["filesPersisted"];
	        this.filesSkipped = source["filesSkipped"];
	        this.errors = source["errors"];
	        this.durationMs = source["durationMs"];
	        this.duplicateGroups = source["duplicateGroups"];
	    }
	}
	export class TidySummary {
	    total: number;
	    moved: number;
	    skipped: number;
	    failed: number;
	    durationMs: number;
	    dryRun: boolean;
	    targetBase: string;
	
	    static createFrom(source: any = {}) {
	        return new TidySummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.moved = source["moved"];
	        this.skipped = source["skipped"];
	        this.failed = source["failed"];
	        this.durationMs = source["durationMs"];
	        this.dryRun = source["dryRun"];
	        this.targetBase = source["targetBase"];
	    }
	}

}

export namespace sql {
	
	export class NullString {
	    String: string;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NullString(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.String = source["String"];
	        this.Valid = source["Valid"];
	    }
	}
	export class NullTime {
	    // Go type: time
	    Time: any;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NullTime(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = this.convertValues(source["Time"], null);
	        this.Valid = source["Valid"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace storage {
	
	export class MediaFile {
	    ID: number;
	    Path: string;
	    HashMD5: string;
	    SizeBytes: number;
	    // Go type: time
	    ModTime: any;
	    TakenAt: sql.NullTime;
	    CameraMake: sql.NullString;
	    CameraModel: sql.NullString;
	    MimeType: sql.NullString;
	
	    static createFrom(source: any = {}) {
	        return new MediaFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Path = source["Path"];
	        this.HashMD5 = source["HashMD5"];
	        this.SizeBytes = source["SizeBytes"];
	        this.ModTime = this.convertValues(source["ModTime"], null);
	        this.TakenAt = this.convertValues(source["TakenAt"], sql.NullTime);
	        this.CameraMake = this.convertValues(source["CameraMake"], sql.NullString);
	        this.CameraModel = this.convertValues(source["CameraModel"], sql.NullString);
	        this.MimeType = this.convertValues(source["MimeType"], sql.NullString);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DuplicateGroup {
	    Hash: string;
	    Files: MediaFile[];
	
	    static createFrom(source: any = {}) {
	        return new DuplicateGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Hash = source["Hash"];
	        this.Files = this.convertValues(source["Files"], MediaFile);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

