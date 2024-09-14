#include "LlamaServer.h"
#include <thread>
#include <chrono>
#include <juce_core/juce_core.h>

const juce::File resourcesDir = juce::File::getSpecialLocation(juce::File::currentApplicationFile)
                                    .getChildFile("Contents")
                                    .getChildFile("Resources");
const juce::String LLAMA_SERVER_PATH = resourcesDir.getChildFile("llama-server").getFullPathName();
const juce::String MODEL_PATH = resourcesDir.getChildFile("gemma-2b-it.fp16.gguf").getFullPathName();

LlamaServer::LlamaServer() {}

bool LlamaServer::startProcess()
{
    bool result = childProcess.start(LLAMA_SERVER_PATH + " -m " + MODEL_PATH);
    std::this_thread::sleep_for(std::chrono::seconds(3));
    return result;
}

void LlamaServer::stopProcess()
{
    childProcess.kill();
}

juce::String LlamaServer::readProcessOutput()
{
    return childProcess.readAllProcessOutput();
}