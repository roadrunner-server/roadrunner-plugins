<?php

/**
 * This file is part of RoadRunner package.
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

declare(strict_types=1);

namespace Spiral\RoadRunner\Metrics;

use Spiral\Goridge\RPC\Exception\ServiceException;
use Spiral\Goridge\RPC\RPCInterface;
use Spiral\RoadRunner\Metrics\Exception\MetricsException;

class Metrics implements MetricsInterface
{
    /**
     * @var string
     */
    private const SERVICE_NAME = 'metrics';

    /**
     * @var RPCInterface
     */
    private RPCInterface $rpc;

    /**
     * @param RPCInterface $rpc
     */
    public function __construct(RPCInterface $rpc)
    {
        $this->rpc = $rpc->withServicePrefix(self::SERVICE_NAME);
    }

    /**
     * {@inheritDoc}
     */
    public function add(string $name, float $value, array $labels = []): void
    {
        try {
            $this->rpc->call('Add', \compact('name', 'value', 'labels'));
        } catch (ServiceException $e) {
            throw new MetricsException($e->getMessage(), (int)$e->getCode(), $e);
        }
    }

    /**
     * {@inheritDoc}
     */
    public function sub(string $name, float $value, array $labels = []): void
    {
        try {
            $this->rpc->call('Sub', \compact('name', 'value', 'labels'));
        } catch (ServiceException $e) {
            throw new MetricsException($e->getMessage(), (int)$e->getCode(), $e);
        }
    }

    /**
     * {@inheritDoc}
     */
    public function observe(string $name, float $value, array $labels = []): void
    {
        try {
            $this->rpc->call('Observe', \compact('name', 'value', 'labels'));
        } catch (ServiceException $e) {
            throw new MetricsException($e->getMessage(), (int)$e->getCode(), $e);
        }
    }

    /**
     * {@inheritDoc}
     */
    public function set(string $name, float $value, array $labels = []): void
    {
        try {
            $this->rpc->call('Set', \compact('name', 'value', 'labels'));
        } catch (ServiceException $e) {
            throw new MetricsException($e->getMessage(), (int)$e->getCode(), $e);
        }
    }

    /**
     * {@inheritDoc}
     */
    public function declare(string $name, CollectorInterface $collector): void
    {
        try {
            $this->rpc->call('Declare', [
                'name'      => $name,
                'collector' => $collector->toArray()
            ]);
        } catch (ServiceException $e) {
            if (\str_contains($e->getMessage(), 'tried to register existing collector')) {
                // suppress duplicate metric error
                return;
            }

            throw new MetricsException($e->getMessage(), (int)$e->getCode(), $e);
        }
    }
}
