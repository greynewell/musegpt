cmake_minimum_required(VERSION 3.22)
project(MUSE_GPT VERSION 0.0.1)
add_subdirectory(JUCE)  

# Setup JUCE build settings, source code and compile definitions

juce_add_plugin(musegpt
    NEEDS_WEB_BROWSER TRUE
    PLUGIN_MANUFACTURER_CODE Grey
    PLUGIN_CODE Muse 
    FORMATS AAX AU AUv3 VST3 Standalone
    PRODUCT_NAME "musegpt"
    COPY_PLUGIN_AFTER_BUILD TRUE)
juce_generate_juce_header(musegpt)

target_sources(musegpt
    PRIVATE
        src/main.cpp
        src/LlamaServer.cpp)

target_compile_definitions(musegpt
    PUBLIC
        JUCE_DEBUG=1  # Enable debug mode
        JUCE_LOGGING=1  # Enable logging
        JUCE_STRICT_REFCOUNTEDPOINTER=1
        JUCE_WEB_BROWSER=1
        JUCE_USE_CURL=0
        JUCE_VST3_CAN_REPLACE_VST2=0)

target_link_libraries(musegpt
    PRIVATE
        juce::juce_audio_basics
        juce::juce_audio_devices
        juce::juce_audio_formats
        juce::juce_audio_plugin_client
        juce::juce_audio_processors
        juce::juce_dsp
        juce::juce_audio_utils
        juce::juce_core
        juce::juce_data_structures
        juce::juce_events
        juce::juce_graphics
        juce::juce_gui_basics
        juce::juce_gui_extra
    PUBLIC
        juce::juce_recommended_config_flags
        juce::juce_recommended_lto_flags
        juce::juce_recommended_warning_flags)

# Debug-specific settings
target_compile_definitions(musegpt
PRIVATE
    $<$<CONFIG:Debug>:DEBUG=1>
    $<$<CONFIG:Debug>:_DEBUG=1>
)

# Release-specific settings
target_compile_definitions(musegpt
PRIVATE
    $<$<CONFIG:Release>:NDEBUG=1>
    $<$<CONFIG:Release>:JUCE_DISABLE_ASSERTIONS=1>
)

# Set optimization flags
target_compile_options(musegpt
PRIVATE
    $<$<CONFIG:Debug>:-O0 -g>
    $<$<CONFIG:Release>:-O3>
)

# Define the executable suffix
set(EXE_SUFFIX "")
if (WIN32)
    set(EXE_SUFFIX ".exe")
endif()

# Add llama-server as a binary resource
add_custom_command(
    OUTPUT ${CMAKE_BINARY_DIR}/llama-server${EXE_SUFFIX}
    COMMAND ${CMAKE_COMMAND} -E copy
            ${CMAKE_SOURCE_DIR}/build/llama.cpp/bin/llama-server${EXE_SUFFIX}
            ${CMAKE_BINARY_DIR}/llama-server${EXE_SUFFIX}
    DEPENDS ${CMAKE_SOURCE_DIR}/build/llama.cpp/bin/llama-server${EXE_SUFFIX}
)
add_custom_target(copy_llama_server ALL DEPENDS ${CMAKE_BINARY_DIR}/llama-server${EXE_SUFFIX})

# Ensure the executable is copied into the bundle
set_source_files_properties(${CMAKE_BINARY_DIR}/llama-server${EXE_SUFFIX} PROPERTIES MACOSX_PACKAGE_LOCATION Resources)
target_sources(musegpt PRIVATE ${CMAKE_BINARY_DIR}/llama-server${EXE_SUFFIX})

# Platform-specific resource copying
if (WIN32)
    # Copy llama-server to Windows output directory
    add_custom_command(TARGET musegpt POST_BUILD
        COMMAND ${CMAKE_COMMAND} -E copy
                ${CMAKE_BINARY_DIR}/llama-server${EXE_SUFFIX}
                $<TARGET_FILE_DIR:musegpt>/llama-server${EXE_SUFFIX}
    )
else()
    # Copy llama-server to macOS or Linux plugin formats
    add_custom_command(TARGET musegpt POST_BUILD
        COMMAND ${CMAKE_COMMAND} -E copy
                ${CMAKE_BINARY_DIR}/llama-server${EXE_SUFFIX}
                $<TARGET_FILE_DIR:musegpt>/VST3/musegpt.vst3/Contents/Resources/llama-server${EXE_SUFFIX}
    )
    add_custom_command(TARGET musegpt POST_BUILD
        COMMAND ${CMAKE_COMMAND} -E copy
                ${CMAKE_BINARY_DIR}/llama-server${EXE_SUFFIX}
                $<TARGET_FILE_DIR:musegpt>/Standalone/musegpt.app/Contents/Resources/llama-server${EXE_SUFFIX}
    )
endif()

# Copy model weights to VST plugin format's output directory
add_custom_command(TARGET musegpt POST_BUILD
    COMMAND ${CMAKE_COMMAND} -E copy
            ${CMAKE_SOURCE_DIR}/models/gemma-2b-it.fp16.gguf
            $<TARGET_FILE_DIR:musegpt>/VST3/musegpt.vst3/Contents/Resources/gemma-2b-it.fp16.gguf
)

# Copy model weights to Standalone plugin format's output directory
add_custom_command(TARGET musegpt POST_BUILD
    COMMAND ${CMAKE_COMMAND} -E copy
            ${CMAKE_SOURCE_DIR}/models/gemma-2b-it.fp16.gguf
            $<TARGET_FILE_DIR:musegpt>/Standalone/musegpt.app/Contents/Resources/gemma-2b-it.fp16.gguf
)

# Copy model weights to AAX plugin format's output directory
add_custom_command(TARGET musegpt POST_BUILD
    COMMAND ${CMAKE_COMMAND} -E copy
            ${CMAKE_SOURCE_DIR}/models/gemma-2b-it.fp16.gguf
            $<TARGET_FILE_DIR:musegpt>/AAX/musegpt.aaxplugin/Contents/Resources/gemma-2b-it.fp16.gguf
)

# Copy model weights to AU plugin format's output directory
add_custom_command(TARGET musegpt POST_BUILD
    COMMAND ${CMAKE_COMMAND} -E copy
            ${CMAKE_SOURCE_DIR}/models/gemma-2b-it.fp16.gguf
            $<TARGET_FILE_DIR:musegpt>/AU/musegp.component/Contents/Resources/gemma-2b-it.fp16.gguf
)