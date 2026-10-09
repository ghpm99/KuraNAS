export const shuffleItems = <Item>(items: readonly Item[], random: () => number = Math.random) => {
    const shuffledItems = [...items];
    for (let index = shuffledItems.length - 1; index > 0; index -= 1) {
        const swapIndex = Math.floor(random() * (index + 1));
        [shuffledItems[index], shuffledItems[swapIndex]] = [
            shuffledItems[swapIndex]!,
            shuffledItems[index]!,
        ];
    }
    return shuffledItems;
};
