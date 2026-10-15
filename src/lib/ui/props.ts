/** Resolved setup props: Vue casts declared boolean props to false when omitted. */
export type SetupProps<T> = Readonly<T> & {
    readonly [K in keyof T as T[K] extends boolean | undefined
        ? T[K] extends undefined ? never : K
        : never]-?: boolean;
};
