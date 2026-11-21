/**
 * The part of Common Fabric's guest client this guest uses. The client's
 * source is `base/iframe-sandbox/src/guest.ts` in the labs repository; the
 * build resolves this module to a bundle of it (`tools/vendor-fabric-client.sh`).
 */
declare module '@commonfabric/iframe-sandbox/guest' {
  /** Why the host would not hand the guest a value. */
  export interface CellReadRefusal {
    readonly refusedBy?: string;
    readonly labelSource?: string;
  }

  export interface RemoteCell<T> {
    /** The guest's cached sample; may be stale or undefined. */
    get(): Readonly<T> | undefined;
    /** Demands the host cell's current value. */
    pull(): Promise<T>;
    /** Calls `listener` now with `get()`, then for each later value. */
    sink(
      listener: (value: Readonly<T> | undefined) => void,
      options?: { onRefused?: (refusal: CellReadRefusal) => void },
    ): () => void;
  }

  export interface RemoteStream<T> {
    /** Resolves once the handler's run has committed; rejects if refused. */
    send(event: T): Promise<void>;
  }

  export interface FabricClient {
    cell<T>(name: string): RemoteCell<T>;
    stream<T>(name: string): RemoteStream<T>;
    disconnect(): void;
  }

  export function connectFabric(): FabricClient;
}
