import type { Plugin } from 'unified';

export type Options = Record<string, never>;

const rehypeMathJaxChtmlStub: Plugin = () => {
    return (tree) => tree;
};

export default rehypeMathJaxChtmlStub;
