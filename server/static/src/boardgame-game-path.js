GamePathMixin = function(superClass) {
		return class extends superClass {
		    GamePath(name, id) {
		      return "game/" + name + "/" + id + "/";
		    }
		}
}
