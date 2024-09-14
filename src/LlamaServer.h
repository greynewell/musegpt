#pragma once

#include <JuceHeader.h>

class LlamaServer
{
public:
    LlamaServer();

    bool startProcess();
    void stopProcess();
    juce::String readProcessOutput();

private:
    juce::ChildProcess childProcess;
};