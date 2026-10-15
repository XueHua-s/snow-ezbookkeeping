import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import {
    type CodingAssistantResponse,
    CodingAssistantType
} from '@/models/large_language_model.ts';


// Owns coding for the explorer store.
export function useExplorerCoding() {
    function generateCustomChartCode({ userPrompt, currentCode, cancelableUuid }: { userPrompt: string, currentCode: string, cancelableUuid?: string }): Promise<CodingAssistantResponse> {
        return new Promise((resolve, reject) => {
            services.generateCode({
                type: CodingAssistantType.InsightsExplorerCustomChart,
                userPrompt: userPrompt,
                code: currentCode
            }, cancelableUuid).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to generate code' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                if (error.canceled) {
                    reject(error);
                }

                logger.error('failed to generate code', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to generate code' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function cancelGenerateCustomChartCode(cancelableUuid: string): void {
        services.cancelRequest(cancelableUuid);
    }

    return { generateCustomChartCode, cancelGenerateCustomChartCode };
}
