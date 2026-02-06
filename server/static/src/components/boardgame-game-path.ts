/**
 * Mixin to add GamePath utility method for constructing game URLs.
 * @param superClass - The class to extend with GamePath functionality
 * @returns Extended class with GamePath method
 */
export const GamePathMixin = <T extends new (...args: any[]) => {}>(superClass: T) => {
  return class extends superClass {
    /**
     * Constructs a game path URL from game name and ID
     * @param name - Game type name (e.g., "blackjack", "memory")
     * @param id - Game instance ID
     * @returns Formatted game path (e.g., "game/blackjack/12345/")
     */
    GamePath(name: string, id: string): string {
      return `game/${name}/${id}/`;
    }
  };
};
