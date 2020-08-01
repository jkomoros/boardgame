/*
 * Mix in to add GamePath. This documentation is required for the Polymer tool to not barf.
 * @polymer
 * @mixinFunction
 */
export const GamePathMixin = function(superClass) {
		return class extends superClass {
		    GamePath(name, id) {
		      return "game/" + name + "/" + id + "/";
		    }
		}
}
